package main

import (
	"container/list"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type safeUserMessages struct {
	mu sync.Mutex
	v  map[string]*list.List
}

func (u *safeUserMessages) GetMessages(user string) *list.List {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.v[user]
}

func (u *safeUserMessages) AddMessage(user string, m *Message) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.v[user] == nil {
		u.v[user] = list.New()
	}
	u.v[user].PushBack(*m)
}

func (u *safeUserMessages) PopMessage(user string) Message {
	// pop from front of list
	u.mu.Lock()
	defer u.mu.Unlock()
	l := u.v[user]
	return l.Remove(l.Front()).(Message)
}

type safeLoggedInUsers struct {
	mu sync.Mutex
	v  map[string]bool
}

func (u *safeLoggedInUsers) Get(user string) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.v[user]
}

func (u *safeLoggedInUsers) Add(user string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.v[user] = true
}

type safeAddrUser struct {
	mu sync.Mutex
	v  map[string]string
}

func (u *safeAddrUser) Get(addr string) string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.v[addr]
}

func (u *safeAddrUser) Set(addr, user string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.v[addr] = user
}

type safeUserCurrentRead struct {
	mu sync.Mutex
	v  map[string]*Message
}

func (u *safeUserCurrentRead) Get(user string) *Message {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.v[user]
}

func (u *safeUserCurrentRead) Set(user string, m *Message) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.v[user] = m
}

// store various data structures needed in memory
var loggedInUsers safeLoggedInUsers     // set of logged in users
var addrUser safeAddrUser               // map between client's address to user
var userCurrentRead safeUserCurrentRead // map between user to current read message
var userMessages safeUserMessages       // map between user to list of messages

func Init() {
	// initialize data structures needed
	loggedInUsers = safeLoggedInUsers{v: make(map[string]bool)}
	addrUser = safeAddrUser{v: make(map[string]string)}
	userMessages = safeUserMessages{v: make(map[string]*list.List)}
	userCurrentRead = safeUserCurrentRead{v: make(map[string]*Message)}
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

		// store client's ip address as current logged in user
		addrUser.Set(addr, user)
		// add current user to the set of logged in users
		loggedInUsers.Add(user)

		out = fmt.Sprintf("Logged in as %v", user)

	case "send":
		recipient, msg := commandArr[1], strings.Join(commandArr[2:], " ")
		// check if user is logged in from the IP address
		sender := addrUser.Get(addr)
		if err = checkLogin(sender); err != nil {
			return
		}
		m := &Message{text: msg, sender: sender}
		if err = m.sendMessage(recipient); err != nil {
			return
		}
		out = "message sent"

	case "read":
		user := addrUser.Get(addr)
		if err = checkLogin(user); err != nil {
			return
		}

		// check if there is any message to read
		l := userMessages.GetMessages(user)
		if l == nil || l.Len() == 0 {
			out = "no message to read"
			return
		}
		m := userMessages.PopMessage(user)

		sender := m.sender
		if m.forwarder != nil {
			sender = strings.Join(m.forwarder, ", ") // if this message was forwarded, the forwarders are shown
		}

		out = fmt.Sprintf("from %v: %v", sender, m.text)

		// store current read message for use in reply and forward
		userCurrentRead.Set(user, &m)

	case "reply":
		user := addrUser.Get(addr)
		if err = checkLogin(user); err != nil {
			return
		}
		msg := strings.Join(commandArr[1:], " ")
		readMsg := userCurrentRead.Get(addrUser.Get(addr))
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
		user := addrUser.Get(addr)
		if err = checkLogin(user); err != nil {
			return
		}
		m := userCurrentRead.Get(addrUser.Get(addr))
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
		user := addrUser.Get(addr)
		if err = checkLogin(user); err != nil {
			return
		}
		m := &Message{text: msg, sender: user}
		for recipient, _ := range loggedInUsers.v {
			fmt.Println(recipient)
			if recipient != user {
				go m.sendMessage(recipient)
			}
		}
		out = "message broadcasted"
	}
	return
}

// start of helper functions
func checkLogin(user string) (err error) {
	if user == "" {
		err = errors.New("error: please login first")
	}
	return
}

func (m *Message) sendMessage(recipient string) (err error) {
	if loggedInUsers.Get(recipient) == false {
		err = errors.New(fmt.Sprintf("error: recipient %v not found / not logged in", recipient))
		return
	}
	userMessages.AddMessage(recipient, m)
	return
}
