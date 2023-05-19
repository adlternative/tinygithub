/*
Copyright © 2023 ZheNing Hu <adlternative@gmail.com>
*/
package cmd

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/adlternative/tinygithub/pkg/config"
)

var configFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tinygithub",
	Short: "a tiny github",
	Long:  `a tiny github which can support git code repo service`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().String(config.LogLevel, "info", "log level")
	rootCmd.PersistentFlags().String(config.LogFile, "", "log file")
	rootCmd.PersistentFlags().String(config.GitBinPath, "", "git bin path")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "config.json", "config file")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatalf("viper bind flags failed with %v", err)
	}

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("read config failed with %v", err)
	}

	config.InitLog()
	config.InitGitBinPath()
}
