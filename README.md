# TCP custom protocol ~ KELLY

Run the server:
```bash
go build
./tcp-custom-protocol
```

From another terminal, connect to it:
```bash
nc localhost 6969
```

## Methods

### REG

Register your client:
```bash
REG @johndoe
```

> Close the terminal to unregister it.

### USRS

List all registered clients:
```bash
USRS
```

### MSG

Send a message to another client:
```bash
MSG @mariorossi Hello, world!
```