package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const storeVersion = 1

type fileData struct {
	Version   int      `json:"version"`
	Hostnames []string `json:"hostnames"`
}

// JSONStore persists the hostname list in a versioned JSON file.
type JSONStore struct {
	path string
}

// NewJSONStore creates a store for path. Relative paths are resolved once at
// construction so later working-directory changes cannot redirect the store.
func NewJSONStore(path string) (*JSONStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("hostname store path is empty")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve hostname store path: %w", err)
	}

	return &JSONStore{path: absPath}, nil
}

// NewUserConfigStore creates a store under the current user's config directory.
func NewUserConfigStore() (*JSONStore, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate user config directory: %w", err)
	}

	return NewJSONStore(filepath.Join(configDir, "nemu", "hostnames.json"))
}

// Load reads saved hostnames. A missing file represents an empty list; invalid
// data is returned as an error so it cannot be silently overwritten.
func (s *JSONStore) Load(ctx context.Context) ([]string, error) {
	if err := checkContext(ctx); err != nil {
		return nil, fmt.Errorf("load hostnames: %w", err)
	}

	contents, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read hostname store: %w", err)
	}
	if err := checkContext(ctx); err != nil {
		return nil, fmt.Errorf("load hostnames: %w", err)
	}

	var data fileData
	if err := json.Unmarshal(contents, &data); err != nil {
		return nil, fmt.Errorf("decode hostname store: %w", err)
	}
	if data.Version != storeVersion {
		return nil, fmt.Errorf("unsupported hostname store version %d", data.Version)
	}
	if data.Hostnames == nil {
		return nil, errors.New("hostname store has no hostname list")
	}
	for index, hostname := range data.Hostnames {
		if strings.TrimSpace(hostname) == "" {
			return nil, fmt.Errorf("hostname store entry %d is empty", index)
		}
	}

	return append([]string(nil), data.Hostnames...), nil
}

// Add appends hostname and atomically replaces the store file. Duplicate
// entries are preserved because product requirements do not define deduping.
func (s *JSONStore) Add(ctx context.Context, hostname string) ([]string, error) {
	if strings.TrimSpace(hostname) == "" {
		return nil, errors.New("hostname is empty")
	}
	if err := checkContext(ctx); err != nil {
		return nil, fmt.Errorf("add hostname: %w", err)
	}

	hostnames, err := s.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load hostnames before adding: %w", err)
	}
	hostnames = append(hostnames, hostname)
	if err := s.write(ctx, hostnames); err != nil {
		return nil, fmt.Errorf("save hostname: %w", err)
	}

	return append([]string(nil), hostnames...), nil
}

func (s *JSONStore) write(ctx context.Context, hostnames []string) (returnErr error) {
	if err := checkContext(ctx); err != nil {
		return fmt.Errorf("write hostname store: %w", err)
	}

	directory := filepath.Dir(s.path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create hostname store directory: %w", err)
	}

	contents, err := json.MarshalIndent(fileData{
		Version:   storeVersion,
		Hostnames: hostnames,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode hostname store: %w", err)
	}
	contents = append(contents, '\n')

	file, err := os.CreateTemp(directory, ".hostnames-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary hostname store: %w", err)
	}
	temporaryPath := file.Name()
	fileOpen := true
	renamed := false
	defer func() {
		if fileOpen {
			if closeErr := file.Close(); closeErr != nil {
				returnErr = errors.Join(returnErr, fmt.Errorf("close temporary hostname store: %w", closeErr))
			}
		}
		if !renamed {
			if removeErr := os.Remove(temporaryPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				returnErr = errors.Join(returnErr, fmt.Errorf("remove temporary hostname store: %w", removeErr))
			}
		}
	}()

	if err := file.Chmod(0o600); err != nil {
		return fmt.Errorf("set hostname store permissions: %w", err)
	}
	written, err := file.Write(contents)
	if err != nil {
		return fmt.Errorf("write temporary hostname store: %w", err)
	}
	if written != len(contents) {
		return fmt.Errorf("write temporary hostname store: %w", io.ErrShortWrite)
	}
	if err := checkContext(ctx); err != nil {
		return fmt.Errorf("write hostname store: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync temporary hostname store: %w", err)
	}
	closeErr := file.Close()
	fileOpen = false
	if closeErr != nil {
		return fmt.Errorf("close temporary hostname store: %w", closeErr)
	}
	if err := os.Rename(temporaryPath, s.path); err != nil {
		return fmt.Errorf("replace hostname store: %w", err)
	}
	renamed = true
	return nil
}

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	return ctx.Err()
}
