package main

import (
	"container/list"
	"errors"
	"fmt"
	"strings"
)

// store various data structures needed in memory
var loggedInUsers map[string]bool       // set of logged in users
var addrUser map[string]string          // map between client's address to user
var userMessages map[string]*list.List  // map between user to list of messages
var userCurrentRead map[string]*Message // map between user to current read message

func init() {
	// initialize data structures needed
	addrUser = make(map[string]string)
	userMessages = make(map[string]*list.List)
	loggedInUsers = make(map[string]bool)
	userCurrentRead = make(map[string]*Message)
}

type Message struct {
	text      string
	sender    string
	forwarder []string // if a Message is forwarded, the sender is appended to this slice
}

func RunCommand(commandArr []string, addr string) (out string, err error) {
	switch commandArr[0] {
	case "login":
		user := commandArr[1]
		addrUser[addr] = user
		loggedInUsers[user] = true
		out = fmt.Sprintf("Logged in as %v", user)

	case "send":
		recipient, msg := commandArr[1], strings.Join(commandArr[2:], " ")
		// check if user is logged in from the IP address
		sender := addrUser[addr]
		if err = checkLogin(addrUser[addr]); err != nil {
			return
		}
		m := &Message{text: msg, sender: sender}
		if err = m.sendMessage(recipient); err != nil {
			return
		}
		out = "message sent"

	case "read":
		user := addrUser[addr]
		if err = checkLogin(user); err != nil {
			return
		}

		// check if there is any message to read
		l := userMessages[user]
		if l == nil || l.Len() == 0 {
			out = "no message to read"
			return
		}
		m := l.Remove(l.Front()).(Message)

		sender := m.sender
		if m.forwarder != nil {
			sender = strings.Join(m.forwarder, ", ") // if this message was forwarded, the forwarders are shown
		}

		out = fmt.Sprintf("from %v: %v", sender, m.text)
		userMessages[user] = l

		// store current read message for use in reply and forward
		userCurrentRead[user] = &m

	case "reply":
		user := addrUser[addr]
		if err = checkLogin(user); err != nil {
			return
		}
		msg := commandArr[1]
		readMsg := userCurrentRead[addrUser[addr]]
		if readMsg == nil {
			out = "no current read message to reply."
			return
		}
		recipient := readMsg.sender
		m := Message{text: msg, sender: user, forwarder: nil}
		if err = m.sendMessage(recipient); err != nil {
			return
		}
		out = fmt.Sprintf("message sent to %v", recipient)

	case "forward":
		recipient := commandArr[1]
		user := addrUser[addr]
		if err = checkLogin(user); err != nil {
			return
		}
		m := userCurrentRead[addrUser[addr]]
		if m == nil {
			out = "no current read message to forward."
			return
		}
		if m.forwarder == nil {
			// nil forwarder means the message hasn't been forwarded, so append the original sender of the message
			m.forwarder = append(m.forwarder, m.sender)
		}
		// append current user on each forward
		m.forwarder = append(m.forwarder, user)

		if err = m.sendMessage(recipient); err != nil {
			return
		}
		out = fmt.Sprintf("message forwarded to %v", recipient)
	case "broadcast":
		msg := strings.Join(commandArr[1:], " ")
		user := addrUser[addr]
		if err = checkLogin(user); err != nil {
			return
		}
		m := &Message{text: msg, sender: user}
		for recipient, _ := range loggedInUsers {
			fmt.Println(recipient)
			if recipient != user {
				go m.sendMessage(recipient)
			}
		}
		out = "message broadcasted"
	}
	return
}

func checkLogin(user string) (err error) {
	if user == "" {
		err = errors.New("error: please login first")
	}
	return
}

func (m *Message) sendMessage(recipient string) (err error) {
	if loggedInUsers[recipient] == false {
		err = errors.New(fmt.Sprintf("error: recipient %v not found / not logged in", recipient))
		return
	}
	if userMessages[recipient] == nil {
		userMessages[recipient] = list.New()
	}
	userMessages[recipient].PushBack(*m)
	return
}
