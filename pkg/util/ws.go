package util

import (
	"io"

	"github.com/gorilla/websocket"
)

type WsWrapper struct {
	r io.Reader
	c *websocket.Conn
}

func NewWsWrapper(c *websocket.Conn) *WsWrapper {
	return &WsWrapper{
		c: c,
	}
}

func (c *WsWrapper) Write(p []byte) (int, error) {
	err := c.c.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *WsWrapper) Read(p []byte) (int, error) {
	for {
		if c.r == nil {
			// Advance to next message.
			var err error
			_, c.r, err = c.c.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.r.Read(p)
		if err == io.EOF {
			// At end of message.
			c.r = nil
			if n > 0 {
				return n, nil
			} else {
				// No data read, continue to next message.
				continue
			}
		}
		return n, err
	}
}

func (c *WsWrapper) Close() error {
	return c.c.Close()
}
