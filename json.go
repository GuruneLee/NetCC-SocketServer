package main

import "log"

// JSON is the format for communication between client and server
type JSON struct {
	C    *Client `json:"-"`
	Type string  `json:"type"`
	Data Data    `json:"data"`
}

// Data is embedded in JSON
type Data struct {
	ID         string   `json:"key"`
	IDs        []string `json:"keys,omitempty"`
	Expression string   `json:"expression,omitempty"`
}

// MakeSendData is sendMe, sendOther를 리턴하는 함수
func (c *Client) MakeSendData(rData JSON, recieveType string) (JSON, JSON) {
	var sendMe JSON
	var sendOther JSON
	switch recieveType {
	case "open":
		log.Printf("Received: %s\n", recieveType)
		sendMe, sendOther = func() (JSON, JSON) {
			IDs := c.server.findIDs()
			m := JSON{c, "welcome", Data{c.id, IDs, ""}}
			o := JSON{c, "enter", Data{c.id, nil, ""}}
			return m, o
		}()
	case "close":
		log.Printf("Received: %s\n", recieveType)
		sendMe, sendOther = func() (JSON, JSON) {
			m := JSON{c, "bye", Data{c.id, nil, ""}}
			o := JSON{c, "exit", Data{c.id, nil, ""}}
			return m, o
		}()
	case "exp":
		log.Printf("Received: %s\n", recieveType)
		sendMe, sendOther = func() (JSON, JSON) {
			m := JSON{c, "exp", Data{c.id, nil, rData.Data.Expression}}
			o := JSON{c, "exp", Data{c.id, nil, rData.Data.Expression}}
			return m, o
		}()
	}

	return sendMe, sendOther
}
