package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/zeelrupapara/trade-engine/config"
	database "gitlab.com/zeelrupapara/trade-engine/database/migrations"
	grpc_server "gitlab.com/zeelrupapara/trade-engine/pkg/grpc"
	"gitlab.com/zeelrupapara/trade-engine/pkg/nats"
	"go.uber.org/zap"
)

// GetAPICommandDef runs app
func GetAPICommandDef(cfg config.AppConfig, logger *zap.Logger) cobra.Command {
	apiCommand := cobra.Command{
		Use:   "engine",
		Short: "To start grpc trade engine service server",
		Long:  `To start grpc trade engine service server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create DB connection
			db, err := database.Connect(cfg.DB)
			if err != nil {
				logger.Error(err.Error(), zap.Any("Setup", "Init DB Connection"))
				return err
			}

			// Create Message Broker connection
			nats, err := nats.NewMsgBrokerClient(cfg)
			if err != nil {
				logger.Error(err.Error(), zap.Any("Setup", "Init Nats Connection"))
				return err
			}

			// Create grpc server
			rpcServer := grpc_server.NewGRPCServer(cfg.Port, logger, db, nats.Nc)
			if err := rpcServer.Run(); err != nil {
				logger.Error(err.Error())
				return err
			}
			return nil
		},
	}

	return apiCommand
}
