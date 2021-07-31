package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	connHost = "localhost"
	connPort = "8080"
	connType = "tcp"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	conn, err := net.Dial(connType, fmt.Sprintf("%v:%v", connHost, connPort))
	if err != nil {
		// handle error
		log.Fatal("Cannot connect to server")
	}
	defer conn.Close()
	for {
		fmt.Print("$ ")
		var cmdString string
		cmdString, err = reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		runCommand(conn, cmdString)
	}
}

func runCommand(conn net.Conn, str string) {
	str = strings.TrimSuffix(str, "\n")
	arrCommandStr := strings.Fields(str)
	if len(arrCommandStr) == 0 {
		return
	}
	switch arrCommandStr[0] {
	case "exit", "quit":
		os.Exit(0)
	case "login":
		if len(arrCommandStr) < 2 {
			fmt.Println("Usage: login [username]")
			return
		}
		writeToServer(conn, str)
	case "send":
		if len(arrCommandStr) < 3 {
			fmt.Println("Usage: send [username] [message]")
			return
		}
		writeToServer(conn, str)
	case "read":
		writeToServer(conn, str)
	case "forward":
		if len(arrCommandStr) < 2 {
			fmt.Println("Usage: forward [username]")
			return
		}
		writeToServer(conn, str)
	case "reply":
		if len(arrCommandStr) < 2 {
			fmt.Println("Usage: reply [message]")
			return
		}
		writeToServer(conn, str)
	case "broadcast":
		if len(arrCommandStr) < 2 {
			fmt.Println("Usage: broadcast [message]")
			return
		}
		writeToServer(conn, str)
	default:
		fmt.Println("Unrecognized command")
	}
}

func writeToServer(conn net.Conn, str string) error {
	// Write to server
	_, err := conn.Write([]byte(str + "\n"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	var res string
	// Wait for server response
	if res, err = bufio.NewReader(conn).ReadString('\n'); err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}
