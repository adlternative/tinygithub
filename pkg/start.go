package tinygithub

import (
	"github.com/adlternative/tinygithub/pkg/config"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/router"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/spf13/viper"
)

func Run() error {
	dbEngine := model.NewDBEngine()

	err := dbEngine.WithUserName(viper.GetString(config.DBUser)).
		WithPassword(viper.GetString(config.DBPassword)).
		WithIp(viper.GetString(config.DBIp)).
		WithPort(viper.GetString(config.DBPort)).
		WithDBName(viper.GetString(config.DBName)).Run()
	if err != nil {
		return err
	}

	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	return router.Run(store)
}
