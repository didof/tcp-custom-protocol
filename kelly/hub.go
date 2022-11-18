package kelly

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type hub struct {
	clients      map[string]*client
	commands     chan command
	registerCh   chan *client
	unregisterCh chan *client
}

func NewHub() *hub {
	return &hub{
		clients:      make(map[string]*client),
		commands:     make(chan command),
		registerCh:   make(chan *client),
		unregisterCh: make(chan *client),
	}
}

func (h *hub) NewClient(conn net.Conn) *client {
	return newClient(conn, h.commands, h.registerCh, h.unregisterCh)
}

func (h *hub) Run() {
	for {
		select {
		case client := <-h.registerCh:
			h.register(client)
		case client := <-h.unregisterCh:
			h.unregister(client)
		case cmd := <-h.commands:
			switch cmd.id {
			case USRS:
				h.listClients(cmd)
			case MSG:
				h.routeMessage(cmd)
			}
		}
	}
}

func (h *hub) register(client *client) {
	if _, exists := h.clients[client.username]; exists {
		client.username = ""
		client.err(errors.New("username taken"))
		return
	}

	h.clients[client.username] = client
	client.write("OK\n")
	log.Printf("client registered: %s\n", client.username)
}

func (h *hub) unregister(client *client) {
	if _, exists := h.clients[client.username]; exists {
		delete(h.clients, client.username)

		log.Printf("client unregistered: %s\n", client.username)
	}
}

func (h *hub) listClients(cmd command) {
	var names []string

	for name := range h.clients {
		names = append(names, name)
	}

	cmd.sender.write(fmt.Sprintf("%s\n", strings.Join(names, ", ")))
}

func (h *hub) routeMessage(cmd command) {
	if cmd.sender.username == "" {
		cmd.sender.err(fmt.Errorf("to send a message you first need to register yourself. Try: REG @johndoe"))
		return
	}

	if user, ok := h.clients[cmd.recipient]; ok {
		if user.username == cmd.sender.username {
			cmd.sender.err(fmt.Errorf("you cannot send a message to yourself. Find other users, try: USRS"))
			return
		}

		user.write(fmt.Sprintf("%s: %s\n", cmd.sender.username, cmd.body))
	} else {
		cmd.sender.err(fmt.Errorf("the user %s is not connected", cmd.recipient))
	}
}
