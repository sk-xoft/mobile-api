package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestRotationRetainsFilesAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api.log")
	w, err := newRotatingFile(path, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"aaaa", "bbbb", "cccc", "dddd"} {
		if _, err := w.Write([]byte(text)); err != nil {
			t.Fatal(err)
		}
	}
	for suffix, want := range map[string]string{"": "dddd", ".1": "cccc", ".2": "bbbb"} {
		got, err := os.ReadFile(path + suffix)
		if err != nil || string(got) != want {
			t.Fatalf("%s: got %q, err %v", suffix, got, err)
		}
	}
	if _, err := newRotatingFile(path, 4); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "dddd" {
		t.Fatal("restart truncated logs")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions: %v", info.Mode())
	}
}

func TestConcurrentFileWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api.log")
	w, err := newRotatingFile(path, 10000)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := w.Write([]byte("event\n")); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	got, err := os.ReadFile(path)
	if err != nil || len(got) != 120 {
		t.Fatalf("bytes=%d, err=%v", len(got), err)
	}
}

func TestLoggerWritesJSONToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api.log")
	t.Setenv("LOG_FILE", path)
	logger, err := New("Asia/Bangkok")
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("persistent event")
	got, err := os.ReadFile(path)
	if err != nil || !json.Valid(got) {
		t.Fatalf("invalid JSON: %q, err=%v", got, err)
	}
	t.Setenv("LOG_FILE", filepath.Join(t.TempDir(), "missing", "api.log"))
	if _, err := New("Asia/Bangkok"); err == nil {
		t.Fatal("expected inaccessible file error")
	}
}
