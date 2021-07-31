package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

var port string

func initServer() {
	// Start the new server
	randomPort := rand.Intn(10000)
	port = fmt.Sprintf(":%d", 10000+randomPort)
	server := NewServer(port)

	// Run the servers in goroutines to stop blocking
	go func() {
		server.Run()
	}()
}

func flushData() {
	// flush all data structures used in command.go
	Init()
}

func TestMain(m *testing.M) {
	initServer()
	// slight delay to make sure server has started before running tests
	time.Sleep(500 * time.Millisecond)
	code := m.Run()
	os.Exit(code)
}

func TestServerRunning(t *testing.T) {
	conn, err := net.Dial("tcp", port)
	if err != nil {
		t.Error("could not connect to server: ", err)
	}
	defer conn.Close()
}

func connectToServer(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", port)
	if err != nil {
		t.Fatalf("could not connect to server %v", err)
	}
	return conn
}

func writeToServer(t *testing.T, conn net.Conn, str string) string {
	// Write to server
	_, err := conn.Write([]byte(str + "\n"))
	if err != nil {
		t.Error("Could not write payload to server ", err)
		return ""
	}

	var res string
	// Wait for server response
	if res, err = bufio.NewReader(conn).ReadString('\n'); err != nil {
		t.Error("Could not receive response from server ", err)
		return ""
	}
	fmt.Println(res)
	return res
}

func TestLogin(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	var res string
	res = writeToServer(t, conn, "login asd")
	if res != "Logged in as asd\n" {
		t.Error("Response not matched")
	}
}

func TestSendWithoutLogin(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	res := writeToServer(t, conn, "send asd lalala")
	if strings.HasPrefix(res, "error") == false {
		t.Error("Should error on send without login first")
	}
}

func TestSendToNonExistingUser(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	res := writeToServer(t, conn, "login asd")
	res = writeToServer(t, conn, "send qwe lalala")
	if strings.HasPrefix(res, "error") == false {
		t.Error("Should error on sending message to non existing user")
	}
}

func TestSendSuccess(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	writeToServer(t, conn, "login asd")
	writeToServer(t, conn, "login qwe")
	res := writeToServer(t, conn, "send asd \"this is from testing\"")
	if res != "message sent\n" {
		t.Error("message not sent")
	}
}

func TestSendConcurrent(t *testing.T) {
	conn := connectToServer(t)
	defer conn.Close()
	writeToServer(t, conn, "login asd")
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		// simulate 10 different clients sending msg to same user at the same time
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := connectToServer(t)
			writeToServer(t, conn, fmt.Sprintf("login %v", i))
			writeToServer(t, conn, fmt.Sprintf("send asd %v", i))
		}(i)
	}
	wg.Wait()

	messages := map[string]bool{}
	for i := 0; i < 10; i++ {
		res := writeToServer(t, conn, "read")
		messages[res] = true
	}

	for i := 0; i < 10; i++ {
		if messages[fmt.Sprintf("from %v: %v\n", i, i)] == false {
			t.Errorf("Message from counter %v not found", i)
		}
	}
}

func TestReadWithoutLogin(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	res := writeToServer(t, conn, "read")
	if strings.HasPrefix(res, "error") == false {
		t.Error("Should have error on read before login")
	}
}

func TestReadSuccess(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	writeToServer(t, conn, "login asd")
	writeToServer(t, conn, "login qwe")
	writeToServer(t, conn, "send asd \"this is from testing\"")
	writeToServer(t, conn, "login asd")
	res := writeToServer(t, conn, "read")
	want := "from qwe: \"this is from testing\"\n"
	if res != want {
		t.Errorf("want: %v - received: %v", want, res)
	}
}

func TestReplyWithoutReadingMessage(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	writeToServer(t, conn, "login asd")
	res := writeToServer(t, conn, "reply test")
	if res != "no current read message to reply.\n" {
		t.Error("Should have no message to reply")
	}
}

func TestReplySuccess(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	writeToServer(t, conn, "login asd")
	writeToServer(t, conn, "login qwe")
	writeToServer(t, conn, "send asd \"this is from testing\"")
	writeToServer(t, conn, "login asd")
	writeToServer(t, conn, "read")
	res := writeToServer(t, conn, "reply \"this is reply from testing\"")
	want := "message sent to qwe\n"
	if res != want {
		t.Errorf("want: %v - received: %v", want, res)
	}

	writeToServer(t, conn, "login qwe")
	res = writeToServer(t, conn, "read")
	want = "from asd: \"this is reply from testing\"\n"
	if res != want {
		t.Errorf("want: %v - received: %v", want, res)
	}
}

func TestForward(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	writeToServer(t, conn, "login asd")
	writeToServer(t, conn, "login qwe")
	writeToServer(t, conn, "send asd \"this is from testing\"")
	writeToServer(t, conn, "login asd")
	writeToServer(t, conn, "read")
	res := writeToServer(t, conn, "forward qwe")
	want := "message forwarded to qwe\n"
	if res != want {
		t.Errorf("want: %v - received: %v", want, res)
	}

	writeToServer(t, conn, "login qwe")
	res = writeToServer(t, conn, "read")
	want = "from qwe, asd: \"this is from testing\"\n"
	if res != want {
		t.Errorf("want: %v - received: %v", want, res)
	}

	// forward again
	writeToServer(t, conn, "login zxc")
	writeToServer(t, conn, "login qwe")
	writeToServer(t, conn, "forward zxc")
	writeToServer(t, conn, "login zxc")
	res = writeToServer(t, conn, "read")
	want = "from qwe, asd, qwe: \"this is from testing\"\n"
	if res != want {
		t.Errorf("want: %v - received: %v", want, res)
	}
}

func TestBroadcast(t *testing.T) {
	defer flushData()
	conn := connectToServer(t)
	defer conn.Close()
	users := []string{"asd", "qwe", "zxc", "jkl"}
	for _, user := range users {
		command := fmt.Sprintf("login %v", user)
		writeToServer(t, conn, command)
	}
	writeToServer(t, conn, "login asd")
	res := writeToServer(t, conn, "broadcast \"hello everyone\"")
	want := "message broadcasted\n"
	if res != want {
		t.Errorf("want: %v - received: %v", want, res)
	}

	for _, user := range users[1:] {
		command := fmt.Sprintf("login %v", user)
		writeToServer(t, conn, command)
		res = writeToServer(t, conn, "read")
		want := "from asd: \"hello everyone\"\n"
		if res != want {
			t.Errorf("want: %v - received: %v", want, res)
		}
	}
}
