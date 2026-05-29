package index

import "github.com/YagoSchramm/GoDepot/domain/entity"

type FileIndex interface {
	Get(userID, name string) (entity.File, error)
	Add(file entity.File)
	Remove(userID, name string)
	RemoveByPrefix(userID, namePrefix string)
	ClearByUserID(userID string)
	ListByUserID(userID string) []entity.File
}
