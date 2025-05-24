package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/zeelrupapara/trade-engine/config"
	"go.uber.org/zap"
)

// Init app initialization
func Init(cfg config.AppConfig, logger *zap.Logger) error {
	grpcServerCmd := GetAPICommandDef(cfg, logger)
	migrateCmd := GetMigrationCommandDef(cfg)

	rootCmd := &cobra.Command{Use: "trade-engine"}
	rootCmd.AddCommand(&grpcServerCmd, &migrateCmd)
	return rootCmd.Execute()
}
