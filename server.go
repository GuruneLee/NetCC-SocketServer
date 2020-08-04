// server.go
//
// server structure
// func newServer
// func server.run
package main

import (
	"fmt"
	"math/rand"
	"time"
)

type server struct {
	clients     map[*Client]bool
	identifiers map[string]*Client
	broadcast   chan JSON
	register    chan *Client
	unregister  chan *Client
}

func newServer() *server {
	return &server{
		broadcast:   make(chan JSON),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		identifiers: make(map[string]*Client),
	}
}

func (s *server) run() {
	for {
		select {
		// resgister: server에 clients, identifiers 추가하기
		case client := <-s.register:
			fmt.Println("register")
			s.clients[client] = true
			//식별자 생성
			id := s.makeID()
			fmt.Println(id)
			//식별자 저장
			s.identifiers[id] = client
			client.id = id

			// resgister: server에서 client, identifiers[client.id] 제거하기
		case client := <-s.unregister:
			fmt.Println("unregister")
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				delete(s.identifiers, client.id)
				close(client.send)
			}
		case msg := <-s.broadcast:
			fmt.Println("broadcaset some messages")
			// msg: {conn: , type: , data:{id: , expression: }}
			// conn: &client, nil
			// broadcast에서 읽어온 msg를 선별해서 각자의 client에 뿌림
			for client := range s.clients {
				select {
				case client.send <- msg:
				default: //client.send가 사용불가능 할때, 실행
					close(client.send)
					delete(s.clients, client)
				}
			}
		}
	}
}

// makeId is Identifier Generator
func (s *server) makeID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Println("makeID1")
	b := make([]byte, 6)
	IDs := s.findIDs()

	for stop := false; !stop; {
		fmt.Println("makeID2")
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}

		if len(IDs) != 0 {
			stop = true
			for _, v := range s.findIDs() {
				if v == string(b) {
					stop = false
					break
				}
			}
		} else {
			stop = true
		}
	}
	fmt.Println("makeID3")

	return string(b)
}

// findIDs is generating ID array\
// used in c.MakeSendData() and s.makeID()
func (s *server) findIDs() []string {
	var r []string
	for k := range s.identifiers {
		r = append(r, k)
	}

	return r
}
