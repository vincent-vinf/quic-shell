package manager

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/vincent-vinf/quic-shell/pkg/types"
)

type ClientManager struct {
	clientMap map[string]*Client
	lock      sync.Mutex
}

func New() *ClientManager {
	return &ClientManager{
		clientMap: make(map[string]*Client),
	}
}

func (m *ClientManager) Handle(ctx context.Context, conn quic.Connection) error {
	rCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	message, err := conn.ReceiveMessage(rCtx)
	if err != nil {
		return err
	}
	msg, err := types.UnpackMsg(message)
	if err != nil {
		return err
	}

	m.addClient(msg.ID, &Client{
		id:   msg.ID,
		conn: conn,
	})
	log.Println("new client", msg.ID)

	return nil
}

func (m *ClientManager) GetClient(id string) *Client {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.clientMap[id]
}

func (m *ClientManager) addClient(id string, client *Client) {
	m.lock.Lock()
	m.clientMap[id] = client
	m.lock.Unlock()
}

type Client struct {
	id   string
	conn quic.Connection
}

func (c *Client) NewStream() (io.ReadWriteCloser, error) {
	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}

	return stream, nil
}
