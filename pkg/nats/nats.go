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
	const maxAttempts = 3
	for attempts := 0; attempts < maxAttempts; attempts++ {
		if natsErr == nats.ErrSlowConsumer {
			_, _, err := sub.Pending()
			if err != nil {
				return
			}
			// Log error, notify operations...
			return
		}

		if natsErr == nil {
			// reconnected, so return
			return
		}

		// other errors, log them and wait a bit
		fmt.Printf("NATS error: %v, attempt %d/%d\n", natsErr, attempts+1, maxAttempts)
		time.Sleep(1 * time.Second)
	}
	// Log error, notify operations...
}

// Connecting with nats
func NewMsgBroker(cfg config.AppConfig) (*MsgBroker, error) {
	url := fmt.Sprint(cfg.Nats.Host, ":", cfg.Nats.Port)
	fmt.Printf("Connecting to Nats  on %s \n", url)

	nc, err := nats.Connect(url, nats.Timeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	nc.SetErrorHandler(natsErrHandler)

	fmt.Println("Connected to Nats")

	return &MsgBroker{Nc: nc}, nil
}
