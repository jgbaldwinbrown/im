package main

import (
	"flag"
	"io"
	"encoding/json"
        "fmt"
        "net"
        "os"
)

type Client struct {
	net.Conn
	Enc *json.Encoder
	Dec *json.Decoder
	Name string
	ErrChan chan error
}

func (c *Client) WriteMessage(m Message) error {
	return c.Enc.Encode(m)
}

type Message struct {
	ClientName string
	Message string
	Starter bool
	Ender bool
}

type MessageWriter interface {
	WriteMessage(msg Message) error
}

func HandleClient(clientConn net.Conn, servWrite chan<- Message) (*Client, error) {
	c := new(Client)
	c.Conn = clientConn
	c.ErrChan = make(chan error)
	c.Enc = json.NewEncoder(c.Conn)
	c.Dec = json.NewDecoder(c.Conn)

	var firstMsg Message
	err := c.Dec.Decode(&firstMsg)
	if err != nil {
		return nil, err
	}
	c.Name = firstMsg.ClientName

	go func() {
		var err error
		defer func() {
			c.ErrChan <- err
			close(c.ErrChan)
		}()

		var inmsg Message
		for e := c.Dec.Decode(&inmsg); e != io.EOF; e = c.Dec.Decode(&inmsg) {
			if e != nil {
				err = e
				return
			}

			servWrite <- inmsg
		}
	}()

	return c, nil
}

type NewConn struct {
	CoreConn io.ReadWriteCloser
	Name string
}

func WriteMsgToClients(msg Message, clients map[string]*Client) error {
	for _, client := range clients {
		e := client.WriteMessage(msg)
		if e != nil {
			return e
		}
	}
	return nil
}

func HandleCore(inChan <-chan Message, newConnChan <-chan *Client) error {
	clients := map[string]*Client{}
	for {
		select {
		case newconn := <-newConnChan:
			clients[newconn.Name] = newconn
		case msg := <-inChan:
			err := WriteMsgToClients(msg, clients)
			if err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func main() {
	listenp := flag.String("l", "localhost:9988", "Address to listen to")
	flag.Parse()

        fmt.Println("Server Running...")
        server, err := net.Listen("tcp", *listenp)
        if err != nil {
                fmt.Println("Error listening:", err.Error())
                os.Exit(1)
        }
        defer server.Close()
        fmt.Println("Listening on localhost:9988")
        fmt.Println("Waiting for client...")

	clientChan := make(chan *Client)
	serverChan := make(chan Message)

	go func() {
		for {
			connection, err := server.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			fmt.Println("client connected")

			client, err := HandleClient(connection, serverChan)
			if err != nil {
				panic(err)
			}
			clientChan <- client
		}
	}()

	err = HandleCore(serverChan, clientChan)
	if err != nil {
		panic(err)
	}
}
