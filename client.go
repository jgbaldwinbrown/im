// socket-client project main.go
package main

import (
	"io"
	"bufio"
	"encoding/json"
        "fmt"
        "net"
)

type Message struct {
	ClientName string
	Message string
	Starter bool
	Ender bool
}

func WriteToServer(name string, term io.Reader, serv io.WriteCloser) error {
	defer serv.Close()
	s := bufio.NewScanner(term)
	enc := json.NewEncoder(serv)

	e := enc.Encode(Message{ClientName: name, Starter: true})
	if e != nil { return e }

	for s.Scan() {
		if s.Err() != nil { return s.Err() }
		e = enc.Encode(Message{ClientName: name, Message: s.Text()})
		if e != nil { return e }
	}

	return nil
}

func ReadFromServer(term io.Writer, serv io.Reader) error {
	var msg Message
	dec := json.NewDecoder(serv)

	for e := dec.Decode(&msg); e != io.EOF; e = dec.Decode(&msg) {
		if e != nil { return e }
		_, e := fmt.Fprintf(term, "\n%v:\n%v\n", msg.ClientName, msg.Message)
		if e != nil { return e }
	}

	return nil
}

func HandleConn(conn io.ReadWriteCloser, term io.ReadWriter, name string) error {
	ec := make(chan error)
	go func() {
		ec <- ReadFromServer(term, conn)
	}()
	go func() {
		ec <- WriteToServer(name, term, conn)
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

func main() {
        //establish connection
        conn, err := net.Dial("tcp", "localhost:9988")
        if err != nil {
                panic(err)
        }
        defer conn.Close()

        ///send some data
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	err = enc.Encode(20)
        if err != nil {
                panic(err)
        }

	var f int
        err = dec.Decode(&f)
        if err != nil {
                fmt.Println("Error reading:", err.Error())
        }
	fmt.Println("f:", f)
}
