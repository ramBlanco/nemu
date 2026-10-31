package tui

import (
	"charm.land/lipgloss/v2"

	"nemu/internal/dns"
)

type styleSet struct {
	title        lipgloss.Style
	resolved     lipgloss.Style
	notResolved  lipgloss.Style
	inconclusive lipgloss.Style
	storageError lipgloss.Style
	muted        lipgloss.Style
}

var styles = styleSet{
	title:        lipgloss.NewStyle().Bold(true),
	resolved:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")),
	notResolved:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
	inconclusive: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208")),
	storageError: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")),
	muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
}

func (s styleSet) forOutcome(status dns.OutcomeStatus) lipgloss.Style {
	switch status {
	case dns.OutcomeResolved:
		return s.resolved
	case dns.OutcomeNotResolved:
		return s.notResolved
	default:
		return s.inconclusive
	}
}

func outcomeLabel(status dns.OutcomeStatus) string {
	switch status {
	case dns.OutcomeResolved:
		return "Resuelve"
	case dns.OutcomeNotResolved:
		return "No resuelve"
	default:
		return "No concluyente"
	}
}
