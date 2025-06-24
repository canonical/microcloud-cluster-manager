package cli

import (
	"fmt"
	"os"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/spf13/cobra"
)

// CmdControl contains common flags for the CLI commands.
type CmdControl struct {
	FlagHelp bool
}

// Run is the main entry point for the CLI tool.
func Run() (err error) {
	// =========================================================================
	// Load configuration

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("Failed to load configuration: %w", err)
	}

	err = cfg.LoadCertificates()
	if err != nil {
		return fmt.Errorf("failed to load certificates: %w", err)
	}

	// =========================================================================
	// Database Support

	dbConfigs := database.DBConfig{
		DBHost:         cfg.DBHost,
		DBUser:         cfg.DBUser,
		DBPassword:     cfg.DBPassword,
		DBName:         cfg.DBName,
		DBMaxIdleConns: cfg.DBMaxIdleConns,
		DBMaxOpenConns: cfg.DBMaxOpenConns,
		DBDisableTLS:   cfg.DBDisableTLS,
	}

	db, err := database.NewDB(dbConfigs)
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}
	defer func() {
		err = db.Close()
	}()

	// =========================================================================
	// Initialize commands

	commonCmd := CmdControl{}
	app := &cobra.Command{
		Use:   "microcloud-cluster-manager",
		Short: "Command for managing the MicroCloud cluster manager",
	}

	app.PersistentFlags().BoolVarP(&commonCmd.FlagHelp, "help", "h", false, "Print help")

	var cmdEnrol = cmdEnrol{
		CFG: cfg,
		DB:  db,
	}
	app.AddCommand(cmdEnrol.Command())

	err = app.Execute()
	if err != nil {
		os.Exit(1)
	}

	return nil
}
