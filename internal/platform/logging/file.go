package logging

import (
	"fmt"
	"os"
	"sync"
)

// rotatingFile keeps the current file and two backups. Opening per write avoids
// retaining descriptors across rotation or shutdown. Only one process may write
// to a given path.
type rotatingFile struct {
	mu    sync.Mutex
	path  string
	limit int64
}

func newRotatingFile(path string, limit int64) (*rotatingFile, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return &rotatingFile{path: path, limit: limit}, nil
}

func (w *rotatingFile) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	info, err := os.Stat(w.path)
	if err != nil {
		return 0, err
	}
	if info.Size() > 0 && info.Size()+int64(len(p)) > w.limit {
		if err := os.Remove(w.path + ".2"); err != nil && !os.IsNotExist(err) {
			return 0, err
		}
		if err := os.Rename(w.path+".1", w.path+".2"); err != nil && !os.IsNotExist(err) {
			return 0, err
		}
		if err := os.Rename(w.path, w.path+".1"); err != nil {
			return 0, err
		}
	}
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}
	n, err := f.Write(p)
	closeErr := f.Close()
	if err != nil {
		return n, err
	}
	return n, closeErr
}
