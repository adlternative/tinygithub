/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/adlternative/tinygithub/pkg"
	"github.com/adlternative/tinygithub/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "tinygithub http server",
	Long:  `support tinygithub http service`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := tinygithub.Run(); err != nil {
			log.Fatalf("tinygithub server failed with: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	serverCmd.PersistentFlags().String(config.Storage, "/app/storage", "git repositories storage path")
	serverCmd.PersistentFlags().String(config.DBUser, "root", "database user")
	serverCmd.PersistentFlags().String(config.DBPassword, "123456", "database password")
	serverCmd.PersistentFlags().String(config.DBIp, "127.0.0.1", "database Ip")
	serverCmd.PersistentFlags().String(config.DBPort, "3306", "database port")
	serverCmd.PersistentFlags().String(config.DBName, "tinygithub", "database name")
	serverCmd.PersistentFlags().Bool(config.DBSync, false, "database sync")
	serverCmd.PersistentFlags().String(config.ServerIp, "127.0.0.1", "server ip")
	serverCmd.PersistentFlags().String(config.ServerPort, "8083", "server port")
	serverCmd.PersistentFlags().String(config.SessionSecret, "secret", "session secret")
	serverCmd.PersistentFlags().String(config.StaticResourcePath, "./static", "static resource path")
	serverCmd.PersistentFlags().String(config.HtmlTemplatePath, "./pkg/template/*", "html template path")
	serverCmd.PersistentFlags().String(config.APIVersion, "v2", "api version")

	if err := viper.BindPFlags(serverCmd.PersistentFlags()); err != nil {
		log.Fatalf("viper bind serverCmd flags failed with %v", err)
	}
}
