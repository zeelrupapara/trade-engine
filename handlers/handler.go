package handlers

import (
	"os"

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
	// Stop channel for shutdown the service
	StopCh chan os.Signal
}

func NewEngineCore(config config.AppConfig, logger *zap.Logger, db *goqu.Database, nats *nats.Conn) *EngineCore {
	ec := EngineCore{
		Config:   config,
		Logger:   logger,
		DB:       db,
		Nats:     nats,
		Exchange: make(map[string]interface{}),
		StopCh:   make(chan os.Signal, 1),
	}
	return &ec
}

func (ec *EngineCore) StartEngine() {
	// Load Exchanges Configs
	exchanges, err := yamlconf.LoadExchangeConfig("exchange.yaml")
	if err != nil {
		ec.Logger.Error("Failed to load exchange config", zap.Error(err))
	}

	// Load Exchanges
	ec.LoadExchanges(exchanges)

	// Load Watchers
	ec.InitWatcher()
	
	<-ec.StopCh
	ec.Logger.Info("Shutting down")
}
