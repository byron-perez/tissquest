package persistence

import "time"

type PersistentObjectState struct {
	IsProxy      bool
	IsPersistent bool
	Timestamp    time.Time
}

type PersistentObject interface {
	Save()
	Retrieve()
	Update()
	Delete()
}
