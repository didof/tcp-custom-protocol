package kelly

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type client struct {
	conn         net.Conn
	outboundCh   chan<- command
	username     string
	registerCh   chan<- *client
	unregisterCh chan<- *client
}

func newClient(conn net.Conn, commandCh chan<- command, registerCh chan<- *client, unregisterCh chan<- *client) *client {
	return &client{
		conn:         conn,
		outboundCh:   commandCh,
		username:     "",
		registerCh:   registerCh,
		unregisterCh: unregisterCh,
	}
}

func (c *client) write(msg string) {
	if _, err := c.conn.Write([]byte(msg)); err != nil {
		log.Fatal(err)
	}
}

func welcome(c *client) {
	c.write("Connected to server\n")
	c.write("\tREG <username> - register your client with an username. It must start with '@'.\n")
	c.write("\tUSRS - list all active clients.\n")
	c.write("\tMSG <username> <payload...> - send a message to a specific user.\n")
	c.write("\tTo close your connection use Ctrl+C.\n")
}

func (c *client) log(msg string) {
	log.Printf("[%s] %s", c.conn.RemoteAddr(), msg)
}

func (c *client) Read() error {
	welcome(c)

	for {
		msg, err := bufio.NewReader(c.conn).ReadBytes('\n')
		if err == io.EOF {
			c.unregisterCh <- c
			return nil
		}

		if err != nil {
			return err
		}

		c.handle(msg)
	}
}

func (c *client) handle(message []byte) {
	splitted := bytes.Split(bytes.TrimSpace(message), []byte(" "))
	cmd := bytes.ToUpper(splitted[0])

	var args [][]byte

	if len(splitted) > 1 {
		for _, arg := range splitted[1:] {
			args = append(args, bytes.TrimSpace(arg))
		}
	}

	switch ID(cmd) {
	case REG:
		if err := c.registerClient(args); err != nil {
			c.err(err)
		}
	case USRS:
		if err := c.listClients(args); err != nil {
			c.err(err)
		}
	case MSG:
		if err := c.sendMessage(args); err != nil {
			c.err(err)
		}
	}
}

func (c *client) registerClient(args [][]byte) error {
	if len(args) != 1 {
		return errors.New("must provide an username. Try: REG @johndoe")
	}

	username := string(args[0])

	if string(username[0]) != "@" {
		return fmt.Errorf("username must start with '@'. Try: REG @%s", username)
	}

	c.username = string(username)
	c.registerCh <- c

	return nil
}

func (c *client) listClients(args [][]byte) error {
	if len(args) > 0 {
		return errors.New("arguments are not expected. Try: USRS")
	}

	c.outboundCh <- command{
		sender: c,
		id:     USRS,
	}

	return nil
}

func (c *client) sendMessage(args [][]byte) error {
	if len(args) < 2 {
		return errors.New("must specify a recipient (@username)")
	}

	recipient := string(args[0])

	if string(recipient[0]) != "@" {
		return errors.New("recipient must be an user (@username)")
	}

	body := bytes.Join(args[1:], []byte(" "))

	if len(body) == 0 {
		return errors.New("body cannot be empty. Try: MSG @johndoe hello, world")
	}

	c.outboundCh <- command{
		id:        MSG,
		recipient: recipient,
		sender:    c,
		body:      body,
	}

	return nil
}

func (c *client) err(e error) {
	c.write(fmt.Sprintf("ERR %v\n", e))
}
