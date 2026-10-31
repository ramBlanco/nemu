package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestJSONStorePersistsHostnamesAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nemu", "hostnames.json")
	store, err := NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore() error = %v", err)
	}

	got, err := store.Add(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if want := []string{"example.com"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Add() = %v, want %v", got, want)
	}

	reopened, err := NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore() for reopened store: %v", err)
	}
	got, err = reopened.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() after reopening: %v", err)
	}
	if want := []string{"example.com"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %v, want %v", got, want)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat() hostname file: %v", err)
		}
		if gotMode := info.Mode().Perm(); gotMode != 0o600 {
			t.Fatalf("hostname file mode = %04o, want 0600", gotMode)
		}
	}
}

func TestJSONStoreMissingFileIsEmpty(t *testing.T) {
	store, err := NewJSONStore(filepath.Join(t.TempDir(), "missing", "hostnames.json"))
	if err != nil {
		t.Fatalf("NewJSONStore() error = %v", err)
	}

	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() missing file: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Load() = %v, want empty list", got)
	}
}

func TestJSONStoreRejectsCorruptFileWithoutOverwriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hostnames.json")
	original := []byte(`{"version":`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatalf("WriteFile() corrupt store: %v", err)
	}
	store, err := NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore() error = %v", err)
	}

	if _, err := store.Add(context.Background(), "example.com"); err == nil {
		t.Fatal("Add() succeeded with corrupt existing data")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() after failed add: %v", err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("store contents changed after failed add: got %q, want %q", got, original)
	}
}

func TestJSONStoreLoadRejectsUnsupportedOrInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "unsupported version", data: `{"version":2,"hostnames":["example.com"]}`},
		{name: "missing list", data: `{"version":1}`},
		{name: "empty hostname", data: `{"version":1,"hostnames":[""]}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "hostnames.json")
			if err := os.WriteFile(path, []byte(test.data), 0o600); err != nil {
				t.Fatalf("WriteFile() store: %v", err)
			}
			store, err := NewJSONStore(path)
			if err != nil {
				t.Fatalf("NewJSONStore() error = %v", err)
			}
			if _, err := store.Load(context.Background()); err == nil {
				t.Fatal("Load() succeeded with invalid store data")
			}
		})
	}
}

func TestJSONStoreHonorsCanceledContext(t *testing.T) {
	store, err := NewJSONStore(filepath.Join(t.TempDir(), "hostnames.json"))
	if err != nil {
		t.Fatalf("NewJSONStore() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := store.Add(ctx, "example.com"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Add() error = %v, want context.Canceled", err)
	}
}

func TestNewJSONStoreRejectsEmptyPath(t *testing.T) {
	if _, err := NewJSONStore(" "); err == nil {
		t.Fatal("NewJSONStore() succeeded with an empty path")
	}
}
