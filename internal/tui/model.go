package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"nemu/internal/dns"
	"nemu/internal/storage"
)

type hostnamesLoadedMsg struct {
	hostnames []string
	err       error
}

type lookupFinishedMsg struct {
	hostname  string
	outcome   dns.Outcome
	hostnames []string
	storeErr  error
}

// Model is the state for the hostname lookup TUI.
type Model struct {
	input       textinput.Model
	resolver    dns.Resolver
	store       storage.HostnameStore
	focusCmd    tea.Cmd
	hostnames   []string
	storeErr    error
	storeLoaded bool
	loading     bool
	queriedHost string
	outcome     dns.Outcome
	hasResult   bool
}

// NewModel creates the TUI with resolver and persistence dependencies.
func NewModel(resolver dns.Resolver, store storage.HostnameStore) Model {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = "ejemplo.com"
	input.SetWidth(40)
	focusCmd := input.Focus()

	return Model{
		input:    input,
		resolver: resolver,
		store:    store,
		focusCmd: focusCmd,
	}
}

// Init focuses the input and loads saved hostnames without blocking the UI.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.focusCmd, m.loadHostnamesCommand())
}

// Update handles keyboard input, terminal resizing, and DNS completion events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		inputWidth := msg.Width - 16
		if inputWidth < 8 {
			inputWidth = 8
		}
		m.input.SetWidth(inputWidth)
		return m, nil

	case hostnamesLoadedMsg:
		m.storeLoaded = true
		m.storeErr = msg.err
		if msg.err == nil {
			m.hostnames = append([]string(nil), msg.hostnames...)
		}
		return m, nil

	case lookupFinishedMsg:
		m.loading = false
		m.queriedHost = msg.hostname
		m.outcome = msg.outcome
		m.storeErr = msg.storeErr
		if msg.storeErr == nil {
			m.hostnames = append([]string(nil), msg.hostnames...)
		}
		m.hasResult = true
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.loading || !m.storeLoaded {
				return m, nil
			}
			hostname := strings.TrimSpace(m.input.Value())
			if hostname == "" {
				return m, nil
			}
			m.loading = true
			m.hasResult = false
			m.queriedHost = hostname
			m.outcome = dns.Outcome{}
			return m, m.addAndLookupCommand(hostname)
		}
		if m.loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) loadHostnamesCommand() tea.Cmd {
	store := m.store
	return func() tea.Msg {
		if store == nil {
			return hostnamesLoadedMsg{err: errors.New("hostname store is not configured")}
		}
		hostnames, err := store.Load(context.Background())
		return hostnamesLoadedMsg{hostnames: hostnames, err: err}
	}
}

func (m Model) addAndLookupCommand(hostname string) tea.Cmd {
	store := m.store
	resolver := m.resolver
	return func() tea.Msg {
		var hostnames []string
		var storeErr error
		if store == nil {
			storeErr = errors.New("hostname store is not configured")
		} else {
			hostnames, storeErr = store.Add(context.Background(), hostname)
			if storeErr != nil {
				storeErr = fmt.Errorf("save hostname: %w", storeErr)
			}
		}

		var outcome dns.Outcome
		if resolver == nil {
			err := errors.New("DNS resolver is not configured")
			outcome = dns.ClassifyOutcome(nil, err)
		} else {
			result, err := resolver.Lookup(context.Background(), hostname)
			outcome = dns.ClassifyOutcome(result.Addresses, err)
		}

		return lookupFinishedMsg{
			hostname:  hostname,
			outcome:   outcome,
			hostnames: hostnames,
			storeErr:  storeErr,
		}
	}
}
