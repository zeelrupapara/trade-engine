package main

import (
	"gitlab.com/zeelrupapara/news-srv/cmd"
	"gitlab.com/zeelrupapara/news-srv/config"
	"gitlab.com/zeelrupapara/news-srv/logger"
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
