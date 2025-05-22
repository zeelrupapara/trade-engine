package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/zeelrupapara/news-srv/config"
	database "gitlab.com/zeelrupapara/news-srv/database/migrations"
	grpc_server "gitlab.com/zeelrupapara/news-srv/pkg/grpc"
	"go.uber.org/zap"
)

// GetAPICommandDef runs app
func GetAPICommandDef(cfg config.AppConfig, logger *zap.Logger) cobra.Command {
	apiCommand := cobra.Command{
		Use:   "grpc-server",
		Short: "To start grpc news service server",
		Long:  `To start grpc news service server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create DB connection
			db, err := database.Connect(cfg.DB)
			if err != nil {
				logger.Error(err.Error())
				return err
			}

			// Create grpc server
			rpcServer := grpc_server.NewGRPCServer(cfg.Port, logger, db)
			if err := rpcServer.Run(); err != nil {
				logger.Error(err.Error())
				return err
			}
			return nil
		},
	}

	return apiCommand
}
