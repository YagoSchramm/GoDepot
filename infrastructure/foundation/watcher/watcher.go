package watcher

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/index"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

type Watcher struct {
	rootDir string
	index   index.FileIndex
	fsw     *fsnotify.Watcher
}

func NewWatcher(rootDir string, idx index.FileIndex) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		rootDir: rootDir,
		index:   idx,
		fsw:     fsw,
	}, nil
}

func (w *Watcher) Start() error {
	if err := w.indexExisting(); err != nil {
		return err
	}

	// Observa cada subpasta de usuário
	entries, err := os.ReadDir(w.rootDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			path := filepath.Join(w.rootDir, e.Name())
			if err := w.fsw.Add(path); err != nil {
				log.Printf("watcher: failed to watch %s: %v", path, err)
			}
		}
	}

	go w.loop()
	return nil
}

func (w *Watcher) Stop() {
	w.fsw.Close()
}

// loop processa eventos do fsnotify indefinidamente.
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
	path := event.Name
	userID := w.extractUserID(path)
	name := filepath.Base(path)

	switch {
	case event.Has(fsnotify.Create) || event.Has(fsnotify.Write):
		time.Sleep(50 * time.Millisecond)

		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return
		}
		file, err := w.buildFile(path, info)
		if err != nil {
			return
		}
		w.index.Add(file)
		log.Printf("watcher: indexed %s (user: %s)", name, userID)

	case event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename):
		w.index.Remove(userID, name)
		log.Printf("watcher: removed %s (user: %s)", name, userID)
	}
}

func (w *Watcher) indexExisting() error {
	return filepath.Walk(w.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		file, err := w.buildFile(path, info)
		if err != nil {
			return nil // ignora arquivos que não conseguiu processar
		}
		w.index.Add(file)
		return nil
	})
}
func (w *Watcher) buildFile(path string, info os.FileInfo) (entity.File, error) {
	userID, err := uuid.Parse(w.extractUserID(path))
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
		Name:       info.Name(),
		Path:       path,
		MimeType:   mimeType,
		Size:       info.Size(),
		ModifiedAt: info.ModTime(),
	}, nil
}
func (w *Watcher) extractUserID(path string) string {
	rel, err := filepath.Rel(w.rootDir, path)
	if err != nil {
		return "unknown"
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) < 2 {
		return "unknown"
	}
	return parts[0]
}
