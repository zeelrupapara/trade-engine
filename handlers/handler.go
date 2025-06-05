package handlers

import (
	"fmt"
	"os"

	"github.com/doug-martin/goqu/v9"
	"github.com/nats-io/nats.go"
	"gitlab.com/zeelrupapara/trade-engine/config"
	yamlconf "gitlab.com/zeelrupapara/trade-engine/config/exchange"
	"gitlab.com/zeelrupapara/trade-engine/models"
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
	ec := EngineCore{Config: config, Logger: logger, DB: db, Nats: nats, Exchange: make(map[string]interface{})}
	return &ec
}

func (ec *EngineCore) StartEngine() {
	// Load Exchanges Configs
	exchanges, err := yamlconf.LoadExchangeConfig("exchange.yaml")
	if err != nil {
		ec.Logger.Error("Failed to load exchange config", zap.Error(err))
	}

	// Load Exchanges
	for _, exchange := range exchanges.Exchanges {
		if exchange.Active {
			ec.Logger.Info(fmt.Sprintf("Loading %s Client", exchange.Name))
			ec.AddExchangeClient(exchange.Name)

			ec.Logger.Info(fmt.Sprintf("Loading %s Symbols", exchange.Name))
			if err := ec.LoadExSymbols(exchange.Name); err != nil {
				ec.Logger.Fatal("Failed to load symbols", zap.String("exchange", exchange.Name), zap.Error(err))
			}
		}
	}

	// Subscribe to symbol events
	subSymbol := models.SubjectEngineSymbol
	_, err = ec.Nats.QueueSubscribe(subSymbol, models.GroupSymbol, ec.SymbolsEventHandler)
	if err != nil {
		ec.Logger.Fatal("NATS subscription failed", zap.String("subject", subSymbol), zap.Error(err))
	}

	<-ec.StopCh
	ec.Logger.Info("Shutting down")
}
