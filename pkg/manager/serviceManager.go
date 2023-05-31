package service_manager

import (
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/storage"
)

type ServiceManager struct {
	store *storage.Storage
	db    *model.DBEngine
}

func New() *ServiceManager {
	return &ServiceManager{}
}

func (sm *ServiceManager) WithStorage(store *storage.Storage) *ServiceManager {
	sm.store = store
	return sm
}

func (sm *ServiceManager) Storage() *storage.Storage {
	return sm.store
}

func (sm *ServiceManager) WithDBEngine(db *model.DBEngine) *ServiceManager {
	sm.db = db
	return sm
}

func (sm *ServiceManager) DBEngine() *model.DBEngine {
	return sm.db
}
