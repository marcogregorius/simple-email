package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Server struct {
	addr   string
	server net.Listener
}

func NewServer(addr string) Server {
	return Server{addr: addr}
}

func (s *Server) Run() (err error) {
	ln, err := net.Listen("tcp", s.addr)
	log.Println("Server running at ", s.addr)
	if err != nil {
		// handle error
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleConnection(conn)
	}
}

func main() {
	srv := NewServer(":8080")
	srv.Run()
}

func handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()

	log.Println("Client", addr, "connected")
	for {
		buffer, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			log.Println("Closing connection")
			conn.Close()
			return
		}
		commandStr := strings.TrimSuffix(string(buffer), "\n")
		log.Println("Received command: ", commandStr)

		commandArr := strings.Split(commandStr, " ")
		var out string
		if out, err = RunCommand(commandArr, addr); err != nil {
			log.Printf("Error occurred on command %v - %v", commandStr, err)
			conn.Write([]byte(err.Error() + "\n"))
		} else {
			conn.Write([]byte(out + "\n"))
		}
	}
}
