package dns

import (
	"errors"
	"net"
	"testing"
)

func TestClassifyOutcome(t *testing.T) {
	var nilDNSError *net.DNSError
	tests := []struct {
		name      string
		addresses []string
		err       error
		want      OutcomeStatus
		wantErr   bool
	}{
		{
			name:      "answers mean resolved",
			addresses: []string{"192.0.2.1"},
			want:      OutcomeResolved,
		},
		{
			name:    "empty completed lookup means no answers",
			want:    OutcomeNotResolved,
			wantErr: false,
		},
		{
			name:    "not found means no answers",
			err:     &net.DNSError{Err: "no such host", IsNotFound: true},
			want:    OutcomeNotResolved,
			wantErr: true,
		},
		{
			name:    "timeout is inconclusive",
			err:     &net.DNSError{Err: "i/o timeout", IsTimeout: true},
			want:    OutcomeInconclusive,
			wantErr: true,
		},
		{
			name:    "generic resolver error is inconclusive",
			err:     errors.New("resolver unavailable"),
			want:    OutcomeInconclusive,
			wantErr: true,
		},
		{
			name:    "typed nil DNS error is inconclusive",
			err:     nilDNSError,
			want:    OutcomeInconclusive,
			wantErr: true,
		},
		{
			name:      "partial answers still establish resolution and retain error",
			addresses: []string{"192.0.2.1"},
			err:       errors.New("partial response"),
			want:      OutcomeResolved,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outcome := ClassifyOutcome(test.addresses, test.err)
			if outcome.Status != test.want {
				t.Fatalf("status = %v, want %v", outcome.Status, test.want)
			}
			if (outcome.Err != nil) != test.wantErr {
				t.Fatalf("error = %v, want error present: %t", outcome.Err, test.wantErr)
			}
			if len(outcome.Addresses) != len(test.addresses) {
				t.Fatalf("addresses = %v, want %v", outcome.Addresses, test.addresses)
			}
		})
	}
}
