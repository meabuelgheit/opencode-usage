package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpen_NonExistentFile(t *testing.T) {
	_, err := Open("/nonexistent/path/opencode.db")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "database not found") {
		t.Errorf("expected 'database not found' error, got: %v", err)
	}
}

func TestOpen_TildeExpansion(t *testing.T) {
	// This validates that the Open function attempts to expand ~/ paths.
	// We pass a path starting with ~/ and expect it to be expanded.
	// Since the file won't exist, we should get a "not found" error
	// that contains the expanded home directory path.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	// Open with a ~/ path that doesn't exist
	_, err = Open("~/nonexistent-opencode-test/db.db")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	// The error should contain the expanded path (starting with home dir)
	expectedPath := filepath.Join(home, "nonexistent-opencode-test", "db.db")
	if !strings.Contains(err.Error(), expectedPath) {
		t.Errorf("expected error to contain expanded path %q, got: %v", expectedPath, err)
	}
}

func TestOpen_NoTildeExpansion(t *testing.T) {
	// When path doesn't start with ~/, no expansion should occur
	_, err := Open("/tmp/nonexistent-opencode-test-db.db")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	if strings.Contains(err.Error(), "/home/") {
		t.Errorf("expected no home directory expansion, got: %v", err)
	}
}

func TestOpen_EmptyPath(t *testing.T) {
	_, err := Open("")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}
