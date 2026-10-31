package tui

import (
	"testing"

	"nemu/internal/dns"
)

func TestOutcomeLabelsMatchApprovedStory(t *testing.T) {
	tests := []struct {
		name   string
		status dns.OutcomeStatus
		want   string
	}{
		{name: "resolved", status: dns.OutcomeResolved, want: "Resuelve"},
		{name: "no answers", status: dns.OutcomeNotResolved, want: "No resuelve"},
		{name: "inconclusive", status: dns.OutcomeInconclusive, want: "No concluyente"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := outcomeLabel(test.status); got != test.want {
				t.Fatalf("outcomeLabel(%v) = %q, want %q", test.status, got, test.want)
			}
		})
	}
}
