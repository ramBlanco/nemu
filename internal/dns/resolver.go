// Package dns adapts Go's system resolver to the application's lookup contract.
package dns

import (
	"context"
	"net"
)

// LookupResult contains the address answers returned by a hostname lookup.
type LookupResult struct {
	Addresses []string
}

// Resolver performs a hostname lookup. Errors are kept separate from empty
// results so callers do not mistake a failed query for authoritative absence.
type Resolver interface {
	Lookup(ctx context.Context, hostname string) (LookupResult, error)
}

// SystemResolver uses the operating system's configured DNS resolver.
// A nil Resolver uses net.DefaultResolver.
type SystemResolver struct {
	Resolver *net.Resolver
}

// Lookup resolves hostname to its IP addresses using the configured resolver.
func (r SystemResolver) Lookup(ctx context.Context, hostname string) (LookupResult, error) {
	resolver := r.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	addresses, err := resolver.LookupHost(ctx, hostname)
	if err != nil {
		return LookupResult{}, err
	}

	return LookupResult{Addresses: addresses}, nil
}
