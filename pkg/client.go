package litws

import (
	"context"
	"fmt"
	"nhooyr.io/websocket"
)

type Client struct {
	lws    *Litws
	wsConn *websocket.Conn

	writeChan chan []byte
}

func newClient(lws *Litws, wsConn *websocket.Conn) *Client {
	return &Client{lws, wsConn, make(chan []byte, 10)}
}

func (c *Client) readLoop() {
	defer c.lws.removeClient(c)
	for {
		_, msg, err := c.wsConn.Read(context.Background())
		if err != nil {
			fmt.Printf("read msg error: %s", err.Error())
			return
		}
		// todo
		fmt.Printf("rcv msg: %s", string(msg))
	}
}

func (c *Client) writeLoop() {
	defer c.lws.removeClient(c)
	for {
		select {
		case msg := <-c.writeChan:
			err := c.wsConn.Write(context.Background(), websocket.MessageText, msg)
			if err != nil {
				fmt.Printf("send msg error: %s", err.Error())
				return
			}
		}
	}
}
