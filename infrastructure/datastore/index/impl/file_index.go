package impl

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/index"
)

func NewFileIndex() index.FileIndex {
	return &fileIndex{
		mu:    sync.RWMutex{},
		files: make(map[string]entity.File),
	}
}

type fileIndex struct {
	mu    sync.RWMutex
	files map[string]entity.File
}

func (f *fileIndex) Add(file entity.File) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files[key(file.UserID.String(), file.Name)] = file
}

func (f *fileIndex) Get(userID, name string) (entity.File, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, ok := f.files[key(userID, name)]
	if !ok {
		return entity.File{}, derr.NotFoundError
	}
	return file, nil
}

func (f *fileIndex) ListByUserID(userID string) []entity.File {
	f.mu.RLock()
	defer f.mu.RUnlock()

	prefix := userID + ":"
	result := []entity.File{}
	for k, f := range f.files {
		if strings.HasPrefix(k, prefix) {
			result = append(result, f)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (f *fileIndex) Remove(userID, name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.files, key(userID, name))
}

func (f *fileIndex) RemoveByPrefix(userID, namePrefix string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	prefix := key(userID, namePrefix)
	childPrefix := prefix + "/"
	for k := range f.files {
		if k == prefix || strings.HasPrefix(k, childPrefix) {
			delete(f.files, k)
		}
	}
}

func (f *fileIndex) ClearByUserID(userID string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	prefix := userID + ":"
	for k := range f.files {
		if strings.HasPrefix(k, prefix) {
			delete(f.files, k)
		}
	}
}

func key(userID, name string) string {
	return fmt.Sprintf("%s:%s", userID, name)
}
