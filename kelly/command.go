package kelly

type ID string

const (
	REG  ID = "REG"
	USRS ID = "USRS"
	MSG  ID = "MSG"
)

type command struct {
	id        ID
	recipient string
	sender    *client
	body      []byte
}
