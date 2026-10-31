package tui

import (
	"context"
	"errors"
	"net"
	"regexp"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"nemu/internal/dns"
	"nemu/internal/storage"
)

type fakeResolver struct {
	calls  []string
	result dns.LookupResult
	err    error
}

type fakeHostnameStore struct {
	hostnames []string
	loadErr   error
	addErr    error
	added     []string
}

var ansiEscape = regexp.MustCompile("\x1b\\[[0-?]*[ -/]*[@-~]")

func plainText(value string) string {
	return ansiEscape.ReplaceAllString(value, "")
}

func (f *fakeResolver) Lookup(_ context.Context, hostname string) (dns.LookupResult, error) {
	f.calls = append(f.calls, hostname)
	return f.result, f.err
}

func (s *fakeHostnameStore) Load(_ context.Context) ([]string, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return slices.Clone(s.hostnames), nil
}

func (s *fakeHostnameStore) Add(_ context.Context, hostname string) ([]string, error) {
	s.added = append(s.added, hostname)
	if s.addErr != nil {
		return nil, s.addErr
	}
	s.hostnames = append(s.hostnames, hostname)
	return slices.Clone(s.hostnames), nil
}

var _ storage.HostnameStore = (*fakeHostnameStore)(nil)

func newLoadedModel(t *testing.T, resolver dns.Resolver, store *fakeHostnameStore) Model {
	t.Helper()
	model := NewModel(resolver, store)
	loadedMessage := model.loadHostnamesCommand()()
	loaded, ok := loadedMessage.(hostnamesLoadedMsg)
	if !ok {
		t.Fatalf("load command returned %T, want hostnamesLoadedMsg", loadedMessage)
	}
	updated, cmd := model.Update(loaded)
	if cmd != nil {
		t.Fatal("handling loaded hostnames unexpectedly started another command")
	}
	model, ok = updated.(Model)
	if !ok {
		t.Fatalf("Update returned unexpected model type %T", updated)
	}
	return model
}

func submitHostname(t *testing.T, model Model, hostname string) (Model, tea.Msg, *fakeResolver) {
	t.Helper()
	resolver, ok := model.resolver.(*fakeResolver)
	if !ok {
		t.Fatal("model does not use the fake resolver")
	}
	model.input.SetValue(hostname)

	updated, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd == nil {
		t.Fatal("confirming the hostname did not start a DNS command")
	}
	model, ok = updated.(Model)
	if !ok {
		t.Fatal("Update returned an unexpected model type")
	}
	return model, cmd(), resolver
}

func completeHostname(t *testing.T, model Model, hostname string) (Model, *fakeResolver) {
	t.Helper()
	updated, msg, resolver := submitHostname(t, model, hostname)
	completed, ok := msg.(lookupFinishedMsg)
	if !ok {
		t.Fatalf("lookup command returned %T, want lookupFinishedMsg", msg)
	}
	if completed.hostname != hostname {
		t.Fatalf("completion hostname = %q, want %q", completed.hostname, hostname)
	}
	final, cmd := updated.Update(completed)
	if cmd != nil {
		t.Fatal("handling a completed lookup unexpectedly started another command")
	}
	model, ok = final.(Model)
	if !ok {
		t.Fatalf("Update returned unexpected model type %T", final)
	}
	return model, resolver
}

func TestConfirmingHostnameStartsLookup(t *testing.T) {
	resolver := &fakeResolver{result: dns.LookupResult{Addresses: []string{"203.0.113.7"}}}
	store := &fakeHostnameStore{}
	model := newLoadedModel(t, resolver, store)
	_, _, calledResolver := submitHostname(t, model, "example.com")

	if len(calledResolver.calls) != 1 || calledResolver.calls[0] != "example.com" {
		t.Fatalf("lookup calls = %v, want [example.com]", calledResolver.calls)
	}
	if !slices.Equal(store.added, []string{"example.com"}) {
		t.Fatalf("stored hostnames = %v, want [example.com]", store.added)
	}
}

func TestCompletedLookupIsAssociatedWithSubmittedHostname(t *testing.T) {
	resolver := &fakeResolver{result: dns.LookupResult{Addresses: []string{"203.0.113.7"}}}
	model, _ := completeHostname(t, newLoadedModel(t, resolver, &fakeHostnameStore{}), "example.com")
	view := plainText(model.View().Content)

	if !strings.Contains(view, `"example.com"`) || !strings.Contains(view, "Estado: Resuelve") || !strings.Contains(view, "203.0.113.7") {
		t.Fatalf("view does not associate the result with the submitted hostname:\n%s", view)
	}
}

func TestLookupOutcomesArePresentedCorrectly(t *testing.T) {
	tests := []struct {
		name       string
		resolver   *fakeResolver
		wantStatus string
		wantDetail string
	}{
		{
			name:       "completed without answers",
			resolver:   &fakeResolver{},
			wantStatus: "Estado: No resuelve",
		},
		{
			name: "not found",
			resolver: &fakeResolver{err: &net.DNSError{
				Err: "no such host", IsNotFound: true,
			}},
			wantStatus: "Estado: No resuelve",
		},
		{
			name: "temporary resolver failure",
			resolver: &fakeResolver{err: &net.DNSError{
				Err: "i/o timeout", IsTimeout: true,
			}},
			wantStatus: "Estado: No concluyente",
			wantDetail: "i/o timeout",
		},
		{
			name:       "generic resolver failure",
			resolver:   &fakeResolver{err: errors.New("temporary resolver failure")},
			wantStatus: "Estado: No concluyente",
			wantDetail: "temporary resolver failure",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model, _ := completeHostname(t, newLoadedModel(t, test.resolver, &fakeHostnameStore{}), "example.com")
			view := plainText(model.View().Content)
			if !strings.Contains(view, test.wantStatus) {
				t.Fatalf("view = %q, want status %q", view, test.wantStatus)
			}
			if !strings.Contains(view, `"example.com"`) {
				t.Fatalf("view does not associate status with hostname:\n%s", view)
			}
			if test.wantDetail != "" && !strings.Contains(view, test.wantDetail) {
				t.Fatalf("view = %q, want detail %q", view, test.wantDetail)
			}
			if strings.Contains(view, "No concluyente") && strings.Contains(view, "No resuelve") {
				t.Fatalf("inconclusive lookup was also labeled as not resolved:\n%s", view)
			}
		})
	}
}

func TestEmptyHostnameDoesNotStartLookup(t *testing.T) {
	resolver := &fakeResolver{}
	store := &fakeHostnameStore{}
	model := newLoadedModel(t, resolver, store)
	updated, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd != nil {
		t.Fatal("empty hostname started a DNS command")
	}
	if len(resolver.calls) != 0 {
		t.Fatalf("lookup calls = %v, want none", resolver.calls)
	}
	if len(store.added) != 0 {
		t.Fatalf("stored hostnames = %v, want none", store.added)
	}
	if _, ok := updated.(Model); !ok {
		t.Fatalf("Update returned unexpected type %T", updated)
	}
}

func TestSavedHostnamesAreVisibleAfterStartupLoad(t *testing.T) {
	store := &fakeHostnameStore{hostnames: []string{"example.org"}}
	model := newLoadedModel(t, &fakeResolver{}, store)
	view := plainText(model.View().Content)

	if !strings.Contains(view, "Hostnames guardados") || !strings.Contains(view, `"example.org"`) {
		t.Fatalf("saved hostname is not visible after startup load:\n%s", view)
	}
}

func TestStorageFailureIsVisibleWithoutLosingDNSOutcome(t *testing.T) {
	store := &fakeHostnameStore{addErr: errors.New("disk is read-only")}
	resolver := &fakeResolver{result: dns.LookupResult{Addresses: []string{"203.0.113.8"}}}
	model := newLoadedModel(t, resolver, store)
	model, _ = completeHostname(t, model, "example.net")
	view := plainText(model.View().Content)

	if !strings.Contains(view, "Estado: Resuelve") || !strings.Contains(view, "No se pudo guardar/cargar:") {
		t.Fatalf("save failure hid the DNS result or storage error:\n%s", view)
	}
	if strings.Contains(view, `• "example.net"`) {
		t.Fatalf("unsaved hostname was shown as persisted:\n%s", view)
	}
}

func TestSubmitWaitsUntilStartupLoadCompletes(t *testing.T) {
	store := &fakeHostnameStore{}
	resolver := &fakeResolver{}
	model := NewModel(resolver, store)
	model.input.SetValue("example.com")

	updated, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd != nil {
		t.Fatal("submit started before saved hostnames finished loading")
	}
	if len(resolver.calls) != 0 || len(store.added) != 0 {
		t.Fatal("submit performed work before startup load completed")
	}
	if _, ok := updated.(Model); !ok {
		t.Fatalf("Update returned unexpected model type %T", updated)
	}
}

func TestStartupStorageFailureIsVisible(t *testing.T) {
	store := &fakeHostnameStore{loadErr: errors.New("invalid store file")}
	model := newLoadedModel(t, &fakeResolver{}, store)
	view := plainText(model.View().Content)

	if !strings.Contains(view, "No se pudo guardar/cargar:") || !strings.Contains(view, "invalid store file") {
		t.Fatalf("startup storage failure was not shown:\n%s", view)
	}
}
