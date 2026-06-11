package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/db"
)

type App struct {
	// services go here
	db *sql.DB
}

var app *App = &App{}

var appLocation = "/var/lib/tysoncloud"

var rootCmd = &cobra.Command{
	Use: "tysoncloud",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := os.Stat(appLocation)
		if os.IsNotExist(err) {
			err := os.MkdirAll(appLocation, 0755)
			if err != nil {
				return err
			}
		}

		db, err := db.NewSqliteStorage()
		if err != nil {
			return err
		}

		app.db = db

		table := `CREATE TABLE IF NOT EXISTS servers (
			id uuid PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			addr TEXT NOT NULL
		)`

		if _, err := app.db.Exec(table); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(loginCmd)
}
