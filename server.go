package main

import (
	"io"
	"encoding/json"
        "fmt"
        "net"
        "os"
)

type Message struct {
	ClientName string
	Message string
	Starter bool
	Ender bool
}

func HandleClient(client io.ReadCloser, servWrite chan<- Message) error {
	defer client.Close()
	ec := make(chan error)

	go func() {
		var err error
		defer func() {
			servWrite <- Message{ClientName: startmsg.ClientName, Ender: true}
			ec <- err
		}()

		enc := json.Encoder(client)

		startmsg, ok := <-servRead
		if !ok {
			return
		}

		for outmsg := range servRead {
			e := enc.Encode(outmsg)
			if e != nil {
				err = e
				return
			}
		}
	}()

	go func() {
		var err error
		defer func() {
			ec <- err
			close(servWrite)
		}

		var inmsg Message
		dec := json.Decoder(client)
		for e := dec.Decode(&inmsg); e != io.EOF; e = dec.Decode(&inmsg) {
			if e != nil {
				err = e
				return
			}

			servWrite <- inmsg
		}
	}()

	var err error
	for i := 0; i < 2; i++ {
		e := <-ec
		if e != nil {
			err = e
		}
	}
	return err
}

type NewConn struct {
	CoreConn io.ReadWriteCloser
	Name string
}

func HandleCore(inchan <-chan Message, newConnChan <-chan Conn) 

func main() {
        fmt.Println("Server Running...")
        server, err := net.Listen("tcp", "localhost:9988")
        if err != nil {
                fmt.Println("Error listening:", err.Error())
                os.Exit(1)
        }
        defer server.Close()
        fmt.Println("Listening on localhost:9988")
        fmt.Println("Waiting for client...")
        for {
                connection, err := server.Accept()
                if err != nil {
                        fmt.Println("Error accepting: ", err.Error())
                        os.Exit(1)
                }
                fmt.Println("client connected")
                go processClient(connection, connection, connection)
        }
}

func processClient(r io.Reader, w io.Writer, c io.Closer) {
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)
	var req int
	for e := dec.Decode(&req); e != io.EOF; e = dec.Decode(&req) {
		if e != nil {
			fmt.Println("Error reading:", e.Error())
		}
		fmt.Printf("Received: %v\n", req)

		f := fibo(req)

		e = enc.Encode(f)
		if e != nil {
			fmt.Println("Error writing: %v", e)
		}
	}
        c.Close()
}
