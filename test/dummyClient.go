// dummyClient.go
// is for avatar update test.
//
// 1. create websocket with socket-server
// 2. send expression request every 5 seconds
// 3. no metter with response messages

package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// JSON is the format for communication between client and server
type JSON struct {
	Type string `json:"type"`
	Data Data   `json:"data,omitempty"`
}

// Data is embedded in JSON
type Data struct {
	ID         string   `json:"key,omitempty"`
	IDs        []string `json:"keys,omitempty"`
	Expression string   `json:"expression,omitempty"`
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	// connect to socket server
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	//send "open" message at first
	myID := testOpen(c)
	// send "close" message closing client with ctrl+c
	testClose(c, myID)
	// send "exp" message every 5 seconds
	exps := []string{"neutral", "happy", "sleepy"}
	order := 1
	for {
		order = order % 3
		time.Sleep(time.Second * 10)
		testExp(c, myID, exps[order])
		order++
	}

} // end of main

func testOpen(c *websocket.Conn) string {
	openJSON := JSON{Type: "open"}
	openJSONBytes, _ := json.Marshal(openJSON)

	c.WriteMessage(websocket.TextMessage, openJSONBytes)
	_, rep, _ := c.ReadMessage()
	var repData JSON
	json.Unmarshal(rep, &repData)
	myID := repData.Data.ID

	return myID
}

func testClose(c *websocket.Conn, myID string) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		//request
		closeJSON := JSON{
			Type: "close",
			Data: Data{
				ID: myID,
			},
		}
		closeJSONBytes, _ := json.Marshal(closeJSON)
		c.WriteMessage(websocket.TextMessage, closeJSONBytes)

		os.Exit(0)
	}()
}

func testExp(c *websocket.Conn, myID string, exp string) {
	expJSON := JSON{
		Type: "exp",
		Data: Data{
			ID:         myID,
			Expression: exp,
		},
	}
	expJSONBytes, _ := json.Marshal(expJSON)
	c.WriteMessage(websocket.TextMessage, expJSONBytes)
}
