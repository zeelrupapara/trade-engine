package cmd

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // for postgres dialect
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"gitlab.com/zeelrupapara/news-srv/config"
	database "gitlab.com/zeelrupapara/news-srv/database/migrations"
)

// GetMigrationCommandDef initialize migration command
func GetMigrationCommandDef(cfg config.AppConfig) cobra.Command {
	migrateCmd := cobra.Command{
		Use:   "migrate [sub command]",
		Short: "To run db migrate",
		Long: `This command is used to run database migration.
	It has up and down sub commands`,
		Args: cobra.MinimumNArgs(1),
	}

	migrateUp := cobra.Command{
		Use:   "up",
		Short: "It will apply migration(s)",
		Long:  `It will run all remaining migration(s)`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostgresMigration(cfg, "UP")
		},
	}

	migrateDown := cobra.Command{
		Use:   "down",
		Short: "It will revert migration(s)",
		Long:  `It will run all remaining migration(s)`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostgresMigration(cfg, "DOWN")
		},
	}
	migrateCmd.AddCommand(&migrateUp, &migrateDown)
	// Migration commands up, down

	return migrateCmd
}

func runPostgresMigration(cfg config.AppConfig, migrationType string) error {
	migrations := migrate.FileMigrationSource{
		Dir: cfg.DB.MigrationDir,
	}

	db, err := sql.Open(database.POSTGRES, fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Db, cfg.DB.QueryString))
	if err != nil {
		return err
	}

	if migrationType == "UP" {
		_, err = migrate.Exec(db, database.POSTGRES, migrations, migrate.Up)
		if err != nil {
			return err
		}
	} else {
		_, err = migrate.Exec(db, database.POSTGRES, migrations, migrate.Down)
		if err != nil {
			return err
		}
	}

	return nil
}
