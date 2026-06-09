//go:build cgo

package storage

import (
	"context"
	"path/filepath"
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

// TestSQLite_BusyTimeoutSet guards the hardening pragma. With SetMaxOpenConns(1)
// an in-process multi-writer test is a tautology (writes serialize on one
// connection), so we assert the pragma value instead — which, with pragmas in
// the DSN, actually regression-guards the setting across reconnects.
func TestSQLite_BusyTimeoutSet(t *testing.T) {
	db := filepath.Join(t.TempDir(), "m.db")
	s, err := NewSQLite(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var ms int
	if err := s.conn.QueryRow("PRAGMA busy_timeout").Scan(&ms); err != nil {
		t.Fatal(err)
	}
	if ms != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", ms)
	}
}

// TestSQLite_InvalidateDanglingSupersededBy documents that invalidating with a
// superseded_by that does not exist must succeed (foreign_keys stays OFF — it
// was never part of the storage hardening decision and turned a previously-OK
// Invalidate into a constraint error).
func TestSQLite_InvalidateDanglingSupersededBy(t *testing.T) {
	db := filepath.Join(t.TempDir(), "m.db")
	s, err := NewSQLite(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	emb := make([]float32, 768)
	mem, err := s.Add(context.Background(), types.Memory{
		Type: "learning", Area: "fk", Content: "x",
	}, emb)
	if err != nil {
		t.Fatal(err)
	}
	nonexistent := int64(999999)
	if err := s.Invalidate(context.Background(), mem.ID, &nonexistent); err != nil {
		t.Fatalf("Invalidate with dangling superseded_by should succeed, got: %v", err)
	}
}
