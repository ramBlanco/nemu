package dns

import (
	"errors"
	"net"
	"slices"
)

// OutcomeStatus is the user-relevant result of a hostname lookup.
type OutcomeStatus uint8

const (
	// OutcomeResolved means the lookup returned at least one address.
	OutcomeResolved OutcomeStatus = iota + 1
	// OutcomeNotResolved means the lookup completed without answers.
	OutcomeNotResolved
	// OutcomeInconclusive means the lookup failed before it could establish an answer.
	OutcomeInconclusive
)

// Outcome contains the classified DNS result and any associated detail.
type Outcome struct {
	Status    OutcomeStatus
	Addresses []string
	Err       error
}

// ClassifyOutcome separates a completed no-answer lookup from resolver failures.
// Any returned address establishes resolution, including a partial result.
func ClassifyOutcome(addresses []string, err error) Outcome {
	result := Outcome{
		Addresses: slices.Clone(addresses),
		Err:       err,
	}

	if len(addresses) > 0 {
		result.Status = OutcomeResolved
		return result
	}

	if err == nil {
		result.Status = OutcomeNotResolved
		return result
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr != nil && dnsErr.IsNotFound {
		result.Status = OutcomeNotResolved
		return result
	}

	result.Status = OutcomeInconclusive
	return result
}
