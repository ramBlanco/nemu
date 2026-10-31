// Package storage persists hostnames for the local Nemu user.
package storage

import "context"

// HostnameStore loads and appends saved hostnames.
type HostnameStore interface {
	Load(ctx context.Context) ([]string, error)
	Add(ctx context.Context, hostname string) ([]string, error)
}
