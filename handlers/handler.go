package handlers

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/nats-io/nats.go"
	"gitlab.com/zeelrupapara/trade-engine/config"
	yamlconf "gitlab.com/zeelrupapara/trade-engine/config/exchange"
	"go.uber.org/zap"
)

type EngineCore struct {
	// Configuratoins of engine
	Config config.AppConfig
	// Logger for tracking and keep logs stremline
	Logger *zap.Logger
	// Database client connection
	DB *goqu.Database
	// Nats message broker connection
	Nats *nats.Conn
	// Exchange Connection Map for multiple exchanges like binance
	Exchange map[string]interface{}
}

func NewEngineCore(config config.AppConfig, logger *zap.Logger, db *goqu.Database) *EngineCore {
	ec := EngineCore{Config: config, Logger: logger, DB: db, Exchange: make(map[string]interface{})}
	return &ec
}

func (ec *EngineCore) StartEngine() error {
	// Load Exchanges Configs
	exchanges, err := yamlconf.LoadExchangeConfig("exchange.yaml")
	if err != nil {
		ec.Logger.Error(err.Error(), zap.Any("Config", "LoadExchangeConfig"))
		return err
	}

	// Load Exchanges Symbols
	for _, exchange := range exchanges.Exchanges {
		if exchange.Active {
			ec.Logger.Info(fmt.Sprintf("Loading %s Client", exchange.Name))
			ec.AddExchangeClient(exchange.Name)
			ec.Logger.Info(fmt.Sprintf("Loading %s Symbols", exchange.Name))
			if err := ec.LoadExSymbols(exchange.Name); err != nil {
				ec.Logger.Error(err.Error(), zap.Any("Config", "LoadExSymbols"))
				return err
			}

		}

	}
	return nil
}
