package dns

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestSystemResolverReturnsDNSFailuresAsErrors(t *testing.T) {
	dialErr := errors.New("blocked by deterministic test resolver")
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(context.Context, string, string) (net.Conn, error) {
			return nil, dialErr
		},
	}

	result, err := (SystemResolver{Resolver: resolver}).Lookup(context.Background(), "example.invalid")
	if err == nil {
		t.Fatal("Lookup returned no error when the injected DNS dialer failed")
	}
	if !strings.Contains(err.Error(), dialErr.Error()) {
		t.Fatalf("Lookup error = %v, want injected dial failure", err)
	}
	if len(result.Addresses) != 0 {
		t.Fatalf("Lookup addresses = %v, want none on error", result.Addresses)
	}
}
