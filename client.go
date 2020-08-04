// client.go
//
// Client structure
// func (client) readPump
// func serveWs
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		} else {
			fmt.Println("read some messages")
		}

		var rData JSON
		json.Unmarshal(message, &rData)
		recieveType := rData.Type

		sendMe, sendOther := c.MakeSendData(rData, recieveType)

		fmt.Println(sendMe)
		fmt.Println(sendOther)
		c.server.broadcast <- sendMe
		c.server.broadcast <- sendOther
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
				log.Fatal(err)
				return
			} else {
				fmt.Println("make next writer w")
			}

			// routing
			var refineSendData []byte
			if c.rightMSG(message) {
				refineSendData, _ = json.Marshal(message)
			}
			_, err = w.Write(refineSendData)
			if err != nil {
				log.Fatal(err)
			} else {
				fmt.Println("Message")
			}
			/*// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			*/
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) rightMSG(msg JSON) bool {
	ans := false
	conn := msg.C
	t := msg.Type
	bar1 := (conn == c) && ((t == "welcome") || (t == "bye"))
	bar2 := (conn != c) && ((t == "enter") || (t == "exit"))

	if (t == "exp") || bar1 || bar2 {
		ans = true
	}

	return ans
}

func serveWs(hub *server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{server: hub, conn: conn, id: "000000", send: make(chan JSON)}
	client.server.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
