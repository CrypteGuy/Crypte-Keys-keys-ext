package client

import (
	"encoding/json"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/pkg/errors"
)

// Client to websocket.
type Client struct {
	url       *url.URL
	conn      *websocket.Conn
	connected bool

	connectMtx sync.Mutex

	keys []*keys.EdX25519Key
}

// New creates a websocket client.
func New(urs string) (*Client, error) {
	url, err := url.Parse(urs)
	if err != nil {
		return nil, err
	}
	return &Client{
		url:  url,
		keys: []*keys.EdX25519Key{},
	}, nil
}

// Register key.
func (c *Client) Register(key *keys.EdX25519Key) {
	logger.Infof("register %s", key.ID())
	c.keys = append(c.keys, key)
	if c.connected {
		if err := c.sendAuth(key); err != nil {
			c.close()
		}
	}
}

// Close ...
func (c *Client) Close() {
	logger.Infof("close")
	if c.connected {
		err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			// Failed to write close message
		}
	}
	c.close()
}

func (c *Client) close() {
	if c.conn != nil {
		c.connectMtx.Lock()
		c.conn.Close()
		c.connected = false
		c.connectMtx.Unlock()
	}
}

func (c *Client) connect() error {
	logger.Infof("connect")
	if c.connected {
		return errors.Errorf("already connected")
	}
	logger.Infof("dial %s", c.url)
	conn, _, err := websocket.DefaultDialer.Dial(c.url.String(), nil)
	if err != nil {
		return errors.Wrapf(err, "failed to dial")
	}
	c.connectMtx.Lock()
	c.conn = conn
	c.connected = true
	c.connectMtx.Unlock()
	return nil
}

// Connect client.
func (c *Client) Connect() error {
	if err := c.connect(); err != nil {
		return err
	}

	for _, key := range c.keys {
		if err := c.sendAuth(key); err != nil {
			return errors.Wrapf(err, "failed to send auth")
		}
	}

	return nil
}

// ReadMessage reads a message.
func (c *Client) ReadMessage() (*api.Message, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	logger.Infof("read message")
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		c.close()
		return nil, err
	}
	var msg api.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) sendAuth(key *keys.EdX25519Key) error {
	logger.Infof("send auth %s", key.ID())
	b := api.GenerateAuth(key, c.url.Hostname())

	if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
		return errors.Wrapf(err, "failed to write message")
	}
	return nil
}