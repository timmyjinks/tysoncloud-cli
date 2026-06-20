package main

import (
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/db"
	"github.com/timmyjinks/tysoncloud-cli/deploy"
	"github.com/timmyjinks/tysoncloud-cli/store"
)

type App struct {
	store *store.SQLStoreService
	sp    *store.SupabaseStoreService
	dsvc  *deploy.DeployService
}

var spURL = os.Getenv("SUPABASE_URL")
var spKey = os.Getenv("SUPABASE_API_KEY")
var kubeConfig = os.Getenv("KUBECONFIG")

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

		target := cmd
		if cmd.Name() == cobra.ShellCompRequestCmd {
			if found, _, err := cmd.Root().Find(args); err == nil {
				target = found
			}
		}

		if target.Annotations["supabase"] == "true" {
			cli, err := db.NewSupabaseStorage(spURL, spKey)
			if err != nil {
				return err
			}
			spSvc := store.NewSupabaseStoreService(cli)
			app.sp = spSvc
		}

		if target.Annotations["database"] == "true" {
			db, err := db.NewSqliteStorage()
			if err != nil {
				return err
			}
			app.store = store.NewSQLStoreService(db)

		}

		if target.Annotations["deploy"] == "true" {
			dsvc := deploy.NewDeployService(kubeConfig)
			app.dsvc = dsvc
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)
	// rootCmd.AddCommand(migrateCmd)
	// rootCmd.AddCommand(logoutCmd)
	// rootCmd.AddCommand(loginCmd)
}
