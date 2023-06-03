package hook_config

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const (
	ServerIp   = "server-ip"
	ServerPort = "server-port"

	LogLevel        = "hook-log-level"
	LogFile         = "hook-log-file"
	PostReceiveMode = "post-receive"

	UserName = "user-name"
	RepoName = "repo-name"
)

func Init() error {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	return InitLog()
}

func InitLog() error {
	if level := viper.GetString(LogLevel); level != "" {
		if parseLevel, err := log.ParseLevel(level); err != nil {
			return fmt.Errorf("parse log level failed: %w", err)
		} else {
			log.SetLevel(parseLevel)
		}
	}

	log.SetOutput(os.Stdout)

	if logFile := viper.GetString(LogFile); logFile != "" {
		logFile, err := homedir.Expand(logFile)
		if err != nil {
			return fmt.Errorf("failed to get abspath of logFile: %w", err)
		}

		dir := filepath.Dir(logFile)

		// 创建文件所在目录
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.WithError(err).Errorf("dir crreate failed")
			return fmt.Errorf("failed to create directory: %w", err)
		}

		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			return fmt.Errorf("open log file %s failed: %v", logFile, err)
		}
		log.SetOutput(file)
	}
	return nil
}
