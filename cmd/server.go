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
	serverCmd.PersistentFlags().String(config.Storage, "", "git repositories storage path")

	if err := viper.BindPFlags(serverCmd.PersistentFlags()); err != nil {
		log.Fatalf("viper bind serverCmd flags failed with %v", err)
	}
}
