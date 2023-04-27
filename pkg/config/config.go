package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

const (
	LogLevel   = "log-level"
	LogFile    = "log-file"
	Storage    = "storage"
	GitBinPath = "git-bin-path"

	DBUser     = "db-user"
	DBPassword = "db-password"
	DBIp       = "db-ip"
	DBPort     = "db-port"
	DBName     = "db-name"

	ServerIp   = "server-ip"
	ServerPort = "server-port"

	SessionSecret = "session-secret"
)

func InitLog() {
	if level := viper.GetString(LogLevel); level != "" {
		if parseLevel, err := log.ParseLevel(level); err != nil {
			log.Fatalf("parse log level failed: %v", err)
		} else {
			log.SetLevel(parseLevel)
		}
	}

	log.SetOutput(os.Stdout)
	if logFile := viper.GetString(LogFile); logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			log.Fatalf("open log file %s failed: %v", logFile, err)
		}
		log.SetOutput(file)
	}
	log.Printf("Init with LogLevel: %v, LogFile: %v", viper.GetString(LogLevel), viper.GetString(LogFile))
}

func InitGitBinPath() {
	gitBinPath := viper.GetString(GitBinPath)
	stat, err := os.Stat(gitBinPath)
	if err != nil || stat.IsDir() {
		if err != nil {
			log.WithError(err).Errorf("init git bin path failed")
		}
		viper.Set(GitBinPath, "/usr/bin/git")
		return
	}
	log.Printf("Init with GitBinPath: %v", viper.GetString(GitBinPath))
}
