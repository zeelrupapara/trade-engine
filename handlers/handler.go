package handlers

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
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
	// Exchange Connection Map for multiple exchanges like binance
	Exchange *models.Exchange
	// Users Symbols map[exchange][USDT] = SymbolSettings
	Symbols map[string][]models.Symbol
}

func NewEngineCore(config config.AppConfig, logger *zap.Logger, db *goqu.Database) *EngineCore {
	exchange := models.Exchange{
		Clients: map[string]interface{}{},
		Symbols: map[string][]string{},
		Active:  false,
	}
	ec := EngineCore{Config: config, Logger: logger, DB: db, Exchange: &exchange}
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
		ec.Logger.Info(fmt.Sprintf("Loading %s Client", exchange.Name))
		ec.AddExchangeClient(exchange.Name, exchange.Active)
		ec.Logger.Info(fmt.Sprintf("Loading %s Symbols", exchange.Name))
		if err := ec.LoadExSymbols(exchange.Name, exchange.Active); err != nil {
			ec.Logger.Error(err.Error(), zap.Any("Config", "LoadExSymbols"))
			return err
		}
	}
	return nil
}
