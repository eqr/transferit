package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	authCmd "github.com/eqr/eqr-auth/cmd"
	authConfig "github.com/eqr/eqr-auth/config"
	"github.com/eqr/transferit/app/cmd"
	"github.com/eqr/transferit/app/config"
	"github.com/eqr/transferit/app/server"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		cmd.SetContext(context.WithValue(ctx, authCmd.ConfigPathKey, ConfigPath))
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.InitConfig(ConfigPath)
		authCfg := authConfig.InitConfig(ConfigPath)
		srv, err := server.New(cfg, authCfg)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Fatal(srv.Start())
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var ConfigPath string

func Build() {
	authCmd.BuildUserManager()
	cmd.BuildFileManager()
	RootCmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", "./config.yml", "path to the configuration file")
	RootCmd.AddCommand(authCmd.UserManagerCmd)
	RootCmd.AddCommand(cmd.FileManager)
}
