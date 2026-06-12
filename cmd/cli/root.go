package main

import (
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/supabase-community/supabase-go"
	"github.com/timmyjinks/tysoncloud-cli/db"
	"github.com/timmyjinks/tysoncloud-cli/store"
)

type App struct {
	store *store.StoreService
	sp    *supabase.Client
}

var spURL = os.Getenv("SUPABASE_URL")
var spKey = os.Getenv("SUPABASE_KEY")

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

		if cmd.Annotations["supabase"] == "true" {
			sp, err := supabase.NewClient(spURL, spKey, &supabase.ClientOptions{})
			if err != nil {
				return err
			}
			app.sp = sp
		}

		if cmd.Annotations["database"] == "true" {
			db, err := db.NewSqliteStorage()
			if err != nil {
				return err
			}
			app.store = store.NewStoreService(db)

		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(loginCmd)
}
