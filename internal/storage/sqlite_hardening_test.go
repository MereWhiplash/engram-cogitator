//go:build cgo

package storage

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

func TestSQLite_WALEnabled(t *testing.T) {
	db := filepath.Join(t.TempDir(), "m.db")
	s, err := NewSQLite(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var mode string
	if err := s.conn.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}
}

func TestSQLite_ConcurrentWritesNoLock(t *testing.T) {
	db := filepath.Join(t.TempDir(), "m.db")
	s, err := NewSQLite(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	emb := make([]float32, 768) // match the vec column dimension (nomic-embed-text)
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Add(context.Background(), types.Memory{
				Type: "learning", Area: "concurrency", Content: "x",
			}, emb)
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent write failed: %v", err)
	}
}
