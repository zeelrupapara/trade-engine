package main

import (
	"gitlab.com/zeelrupapara/trade-engine/cmd"
	"gitlab.com/zeelrupapara/trade-engine/config"
	"gitlab.com/zeelrupapara/trade-engine/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Init config
	cfg := config.GetConfig()

	// Init Logger
	logger, err := logger.NewRootLogger(cfg.Debug, cfg.IsDevelopment)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	// Entry point of command
	err = cmd.Init(cfg, logger)
	if err != nil {
		panic(err)
	}
}
