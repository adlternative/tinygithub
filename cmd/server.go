/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg"
	"github.com/adlternative/tinygithub/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "tinygithub http server",
	Long:  `support tinygithub http service`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return fmt.Errorf("viper bind hookCmd flags failed with %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		config.Init(configFile)

		if err := tinygithub.Run(); err != nil {
			return fmt.Errorf("tinygithub run failed with: %w", err)
		}

		return nil
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
	serverCmd.PersistentFlags().String(config.DBIp, "localhost", "database Ip")
	serverCmd.PersistentFlags().String(config.DBPort, "3306", "database port")
	serverCmd.PersistentFlags().String(config.DBName, "tinygithub", "database name")
	serverCmd.PersistentFlags().Bool(config.DBSync, false, "database sync")
	serverCmd.PersistentFlags().String(config.ServerIp, "localhost", "server ip")
	serverCmd.PersistentFlags().String(config.ServerPort, "8083", "server port")
	serverCmd.PersistentFlags().String(config.SessionSecret, "secret", "session secret")
	serverCmd.PersistentFlags().String(config.StaticResourcePath, "./static", "static resource path")
	serverCmd.PersistentFlags().String(config.HtmlTemplatePath, "./pkg/template/*", "html template path")
	serverCmd.PersistentFlags().String(config.APIVersion, "v2", "api version")
	serverCmd.PersistentFlags().String(config.LogLevel, "info", "log level")
	serverCmd.PersistentFlags().String(config.LogFile, "", "log file")
	serverCmd.PersistentFlags().String(config.GitBinPath, "/usr/bin/git", "git bin path")
	serverCmd.PersistentFlags().StringVar(&configFile, "config", "config.json", "config file")
}
