package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"gitlab.com/zeelrupapara/trade-engine/config"
)

// for sending reciving message to other microservices
// as communication channel
type MsgBroker struct {
	Nc *nats.Conn
}

func natsErrHandler(nc *nats.Conn, sub *nats.Subscription, natsErr error) {
	// fmt.Printf("error: %v\n", natsErr)
	if natsErr == nats.ErrSlowConsumer {
		_, _, err := sub.Pending()
		if err != nil {
			return
		}
		return
		// Log error, notify operations...
	}
	// check for other errors
}

// Connecting with nats
func NewMsgBrokerClient(cfg config.AppConfig) (*MsgBroker, error) {
	url := fmt.Sprint(cfg.Nats.Host, ":", cfg.Nats.Port)
	fmt.Printf("Connecting to Nats  on %s \n", url)

	nc, err := nats.Connect(url, nats.Timeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	nc.SetErrorHandler(natsErrHandler)

	return &MsgBroker{Nc: nc}, nil

}
