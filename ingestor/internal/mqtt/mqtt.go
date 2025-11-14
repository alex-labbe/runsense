package mqtt

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/alex-labbe/runsense/ingestor/internal/config"
	pmqtt "github.com/eclipse/paho.mqtt.golang"
)

type Handler func(topic string, payload []byte)

// wraps paho client and tracks connection health
type Client struct {
	cfg       config.Config
	client    pmqtt.Client
	handler   Handler
	connected atomic.Bool //atomic provides safe concurrent access. OnConnect -> true, OnConnectionLost -> false
}

var firstConnect atomic.Bool

func New(cfg config.Config, handler Handler, onMessage func([]byte)) *Client {
	c := &Client{
		cfg:     cfg,
		handler: handler,
	}

	opts := pmqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", cfg.MQTTHost, cfg.MQTTPort))
	opts.SetClientID(cfg.MQTTClientID)
	opts.SetUsername(cfg.MQTTUsername)
	opts.SetPassword(cfg.MQTTPassword)
	opts.SetKeepAlive(time.Duration(cfg.MQTTKeepAlive) * time.Second)
	opts.SetCleanSession(cfg.MQTTCleanSession)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(time.Duration(5) * time.Second)

	opts.SetDefaultPublishHandler(func(_ pmqtt.Client, msg pmqtt.Message) {
		c.handler(msg.Topic(), msg.Payload())
	})

	opts.OnConnect = func(cl pmqtt.Client) {
		c.connected.Store(true)

		token := cl.Subscribe(c.cfg.MQTTTopics, byte(c.cfg.MQTTQos), nil)
		if token.Wait() && token.Error() != nil {
			// Subscription failed
			c.connected.Store(false)
		}

		if !firstConnect.Swap(true) { // early return for first connection
			return
		}

		//TODO: increment the reconnects metric

	}

	opts.OnConnectionLost = func(cl pmqtt.Client, err error) {
		c.connected.Store(false)

		log.Printf("Mqtt connection lost: %v", err)
	}

	c.client = pmqtt.NewClient(opts)

	return c
}

// called in main ofter db.New and config are loaded
func (c *Client) Start() error {
	token := c.client.Connect()
	if !token.WaitTimeout(time.Duration(10) * time.Second) {
		return fmt.Errorf("mqtt connect timeout")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("mqtt connect error: %w", err)
	}

	return nil // no error if reaches
}

// called after catching sigterm from kubes
func (c *Client) Stop() {
	if c == nil || c.client == nil {
		return // nothing to kill
	}
	c.client.Disconnect(250)
	c.connected.Store(false)
}

// used by /health handler
func (c *Client) Healthy() bool {
	if c == nil || c.client == nil {
		return false
	}
	return c.connected.Load()
}
