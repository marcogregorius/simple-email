# Simple Email / MIRC App
The application consists of two types of process both written in Go.
- ### Server
	TCP Server accepting requests and writing back the response
- ### Client
	TCP Client writing payload to the server.

## To run
- ### Server
	`cd server`
	`go run .`
- ### Client
	On a different terminal session (possible to run multiple clients at the same time):
	`cd client`
	`go run .`
	Start typing the commands like `login asd`, `send qwe "hello world"`

## Testing
Tests are written for the server only.
Run `go test .` inside `server/`directory.
To test for race condition, `go test -race .`
The application is tested for race condition, such as when multiple clients are send message to the same user at the same time.

## Race Condition
Since the data store is in memory, multiple goroutines (possibly threads) from multiple client requests could read/write the same variable at the same time. To prevent race condition, mutex is added to the necessary shared variables between goroutines (ie list of user messages).

A simulation of how race condition would have occurred can be found at `server/server_test.go` at `func TestSendConcurrent`
