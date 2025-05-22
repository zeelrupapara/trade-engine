package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/zeelrupapara/news-srv/config"
	"go.uber.org/zap"
)

// Init app initialization
func Init(cfg config.AppConfig, logger *zap.Logger) error {
	grpcServerCmd := GetAPICommandDef(cfg, logger)
	migrateCmd := GetMigrationCommandDef(cfg)

	rootCmd := &cobra.Command{Use: "news-srv"}
	rootCmd.AddCommand(&grpcServerCmd, &migrateCmd)
	return rootCmd.Execute()
}
