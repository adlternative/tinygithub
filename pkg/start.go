package tinygithub

import (
	"github.com/adlternative/tinygithub/pkg/router"
	"github.com/adlternative/tinygithub/pkg/storage"
)

func Run() error {
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	return router.Run(store)
}
