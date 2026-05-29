package watcher

import (
	"errors"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/cache"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/index"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

type Watcher struct {
	index index.FileIndex
	cache cache.Cache
	fsw   *fsnotify.Watcher

	mu          sync.RWMutex
	userDirs    map[string][]string
	watchedDirs map[string]binding
}

type binding struct {
	userID uuid.UUID
	root   string
}

func NewWatcher(idx index.FileIndex, fileCache cache.Cache) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		index:       idx,
		cache:       fileCache,
		fsw:         fsw,
		userDirs:    make(map[string][]string),
		watchedDirs: make(map[string]binding),
	}, nil
}

func (w *Watcher) Start() {
	go w.loop()
}

func (w *Watcher) Stop() {
	_ = w.fsw.Close()
}

func (w *Watcher) SetRoot(userID uuid.UUID, rootDir string) error {
	if userID == uuid.Nil {
		return errors.New("user id is required")
	}
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return errors.New("sync folder path is required")
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return err
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("sync folder path must be a directory")
	}

	userKey := userID.String()
	w.removeUserWatches(userKey)
	w.index.ClearByUserID(userKey)
	w.invalidateUser(userID)

	if err := w.addDirRecursive(userID, absRoot, absRoot); err != nil {
		return err
	}
	return w.indexExisting(userID, absRoot)
}

func (w *Watcher) removeUserWatches(userID string) {
	w.mu.Lock()
	dirs := append([]string(nil), w.userDirs[userID]...)
	delete(w.userDirs, userID)
	for _, dir := range dirs {
		delete(w.watchedDirs, dir)
	}
	w.mu.Unlock()

	for _, dir := range dirs {
		if err := w.fsw.Remove(dir); err != nil {
			log.Printf("watcher: failed to stop watching %s: %v", dir, err)
		}
	}
}

func (w *Watcher) addDirRecursive(userID uuid.UUID, root string, dir string) error {
	return filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return err
		}
		return w.addDir(userID, root, path)
	})
}

func (w *Watcher) addDir(userID uuid.UUID, root string, dir string) error {
	cleanDir := filepath.Clean(dir)

	w.mu.RLock()
	_, exists := w.watchedDirs[cleanDir]
	w.mu.RUnlock()
	if exists {
		return nil
	}

	if err := w.fsw.Add(cleanDir); err != nil {
		return err
	}

	w.mu.Lock()
	userKey := userID.String()
	w.watchedDirs[cleanDir] = binding{userID: userID, root: root}
	w.userDirs[userKey] = append(w.userDirs[userKey], cleanDir)
	w.mu.Unlock()
	return nil
}

func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	b, ok := w.lookupBinding(event.Name)
	if !ok {
		return
	}

	switch {
	case event.Has(fsnotify.Create) || event.Has(fsnotify.Write):
		time.Sleep(50 * time.Millisecond)

		info, err := os.Stat(event.Name)
		if err != nil {
			return
		}
		if info.IsDir() {
			if event.Has(fsnotify.Create) {
				if err := w.addDirRecursive(b.userID, b.root, event.Name); err != nil {
					log.Printf("watcher: failed to watch %s: %v", event.Name, err)
				}
				if err := w.indexExistingFrom(b.userID, b.root, event.Name); err != nil {
					log.Printf("watcher: failed to index %s: %v", event.Name, err)
				}
				w.invalidateUser(b.userID)
			}
			return
		}
		file, err := w.buildFile(b.userID, b.root, event.Name, info)
		if err != nil {
			return
		}
		w.index.Add(file)
		w.invalidateUser(b.userID)
		log.Printf("watcher: indexed %s (user: %s)", file.Name, b.userID)

	case event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename):
		name, err := relativeName(b.root, event.Name)
		if err != nil {
			return
		}
		w.index.RemoveByPrefix(b.userID.String(), name)
		w.invalidateUser(b.userID)
		log.Printf("watcher: removed %s (user: %s)", name, b.userID)
	}
}

func (w *Watcher) lookupBinding(path string) (binding, bool) {
	dir := filepath.Clean(path)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	for {
		if b, ok := w.watchedDirs[dir]; ok {
			return b, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return binding{}, false
		}
		dir = parent
	}
}

func (w *Watcher) indexExisting(userID uuid.UUID, root string) error {
	return w.indexExistingFrom(userID, root, root)
}

func (w *Watcher) indexExistingFrom(userID uuid.UUID, root string, start string) error {
	return filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		file, err := w.buildFile(userID, root, path, info)
		if err != nil {
			return nil
		}
		w.index.Add(file)
		return nil
	})
}

func (w *Watcher) buildFile(userID uuid.UUID, root string, path string, info os.FileInfo) (entity.File, error) {
	name, err := relativeName(root, path)
	if err != nil {
		return entity.File{}, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return entity.File{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       name,
		Path:       path,
		MimeType:   mimeType,
		Size:       info.Size(),
		ModifiedAt: info.ModTime(),
	}, nil
}

func relativeName(root string, path string) (string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func (w *Watcher) invalidateUser(userID uuid.UUID) {
	w.cache.Invalidate("files:" + userID.String() + ":")
}
