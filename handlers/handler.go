package handlers

import (
	"os"

	"github.com/doug-martin/goqu/v9"
	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats.go"
	"gitlab.com/zeelrupapara/trade-engine/config"
	yamlconf "gitlab.com/zeelrupapara/trade-engine/config/exchange"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/telegram"
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
	// GoCron
	Cron *gocron.Scheduler
	// Telegram Client
	Telegram *telegram.Client
	// Exchange Connection Map for multiple exchanges like binance
	Exchange map[string]interface{}
	// single account loaded once
	Account *models.Account
	// For opened positions
	Positions map[string]*models.Position
	// Stop channel for shutdown the service
	StopCh chan os.Signal
}

func NewEngineCore(config config.AppConfig, logger *zap.Logger, db *goqu.Database, nats *nats.Conn, cron *gocron.Scheduler, telegram *telegram.Client) *EngineCore {
	ec := EngineCore{
		Config:    config,
		Logger:    logger,
		DB:        db,
		Nats:      nats,
		Cron:      cron,
		Telegram:  telegram,
		Exchange:  make(map[string]interface{}),
		StopCh:    make(chan os.Signal, 1),
		Account:   &models.Account{},
		Positions: map[string]*models.Position{},
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
	// Load Accounts
	ec.LoadOrCreateAccountByName("bot")
	// Load positions
	ec.LoadOpenPositions()
	// Load Watchers
	ec.InitWatcher()

	<-ec.StopCh
	ec.Logger.Info("Shutting down")
}
