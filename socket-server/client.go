// client.go
//
// Client structure
// func (client) readPump
// func serveWs
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is...
type Client struct {
	server *server
	conn   *websocket.Conn
	id     string //여섯자리 스트링
	send   chan JSON
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// message 받아오기
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			//if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			//	fmt.Println("client.go/readPump/ReadMessage error, Error: ", "")
			//}
			fmt.Println("client.go/readPump/ReadMessage error, Error: ", "")
			break
		}

		var rData JSON
		json.Unmarshal(message, &rData)
		fmt.Println("\nrecieved Data", rData)
		recieveType := rData.Type

		sendData := c.MakeSendData(rData, recieveType)
		c.server.broadcast <- sendData
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Println("client.go/writePump/NextWriter error, Error: ", err.Error())
				return
			}

			// routing
			var refineSendData []byte
			refineSendData, unregister := c.refineMSG(message)

			//fmt.Println(c, "unre: ", unregister)
			fmt.Printf("To %s :// Message: %s\n", c.id, string(refineSendData))
			_, err = w.Write(refineSendData)
			if err != nil {
				fmt.Println("client.go/writePump/Write error, Error: ", err.Error())
				return
			}

			if unregister {
				c.server.unregister <- c
			}

			if err := w.Close(); err != nil {
				fmt.Println("client.go/writePump/Close error, Error: ", err.Error())
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Println("client.go/writePump/WriteMessage error, Error: ", err.Error())
				return
			}
		}
	}
}

func (c *Client) refineMSG(msg JSON) ([]byte, bool) {
	conn := msg.C
	t := msg.Type

	var ans []byte
	unre := false
	if t != "exp" {
		switch msg.Type {
		case "open":
			if conn == c {
				msg.Type = "welcome"
			} else {
				msg.Type = "enter"
				msg.Data.IDs = nil
			}
		case "close":
			if conn == c {
				msg.Type = "bye"
				unre = true
			} else {
				msg.Type = "exit"
			}
		}
	}

	ans, _ = json.Marshal(msg)

	return ans, unre
}

func serveWs(hub *server, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("client.go/sesrveWs/Upgrade error, Error: ", err.Error())
		return
	}

	client := &Client{server: hub, conn: conn, id: "000000", send: make(chan JSON)}
	client.server.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
