package main

import "fmt"

// JSON is the format for communication between client and server
type JSON struct {
	C    *Client `json:"-"`
	Type string  `json:"type"`
	Data Data    `json:"data,omitempty"`
}

// Data is embedded in JSON
type Data struct {
	ID         string   `json:"key,omitempty"`
	IDs        []string `json:"keys,omitempty"`
	Expression string   `json:"expression,omitempty"`
}

// MakeSendData is sendMe, sendOther를 리턴하는 함수
func (c *Client) MakeSendData(rData JSON, recieveType string) JSON {
	var sendData JSON

	switch recieveType {
	case "open":
		fmt.Printf("Received: %s\n", recieveType)
		sendData = func() JSON {
			IDs := c.server.findIDs()
			m := JSON{c, "open", Data{c.id, IDs, ""}}
			return m
		}()
	case "close":
		fmt.Printf("Received: %s\n", recieveType)
		sendData = func() JSON {
			m := JSON{c, "close", Data{c.id, nil, ""}}
			return m
		}()
	case "exp":
		fmt.Printf("Received: %s\n", recieveType)
		sendData = func() JSON {
			m := JSON{c, "exp", Data{c.id, nil, rData.Data.Expression}}
			return m
		}()
	}

	return sendData
}
