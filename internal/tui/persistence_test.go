package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"nemu/internal/dns"
	"nemu/internal/storage"
)

func TestAddedHostnameIsVisibleAfterTUIRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "hostnames.json")
	store, err := storage.NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore() error = %v", err)
	}

	resolver := &fakeResolver{result: dns.LookupResult{Addresses: []string{"192.0.2.5"}}}
	model := NewModel(resolver, store)
	loadedMsg := model.loadHostnamesCommand()()
	loaded, ok := loadedMsg.(hostnamesLoadedMsg)
	if !ok {
		t.Fatalf("load command returned %T, want hostnamesLoadedMsg", loadedMsg)
	}
	updated, loadCmd := model.Update(loaded)
	if loadCmd != nil {
		t.Fatal("loading persisted hostnames unexpectedly started another command")
	}
	model, ok = updated.(Model)
	if !ok {
		t.Fatalf("Update returned unexpected model type %T", updated)
	}

	model.input.SetValue("example.com")
	updated, addAndLookupCmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if addAndLookupCmd == nil {
		t.Fatal("submitting hostname did not start persistence and DNS work")
	}
	finishedMsg, ok := addAndLookupCmd().(lookupFinishedMsg)
	if !ok {
		t.Fatal("add/lookup command returned an unexpected message")
	}
	updated, finishCmd := updated.Update(finishedMsg)
	if finishCmd != nil {
		t.Fatal("handling add/lookup completion unexpectedly started another command")
	}
	firstSession, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned unexpected model type %T", updated)
	}
	if firstSession.storeErr != nil {
		t.Fatalf("hostname was not persisted: %v", firstSession.storeErr)
	}

	reopenedStore, err := storage.NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore() after restart: %v", err)
	}
	reopenedModel := NewModel(&fakeResolver{}, reopenedStore)
	reopenedMsg := reopenedModel.loadHostnamesCommand()()
	reopened, ok := reopenedMsg.(hostnamesLoadedMsg)
	if !ok {
		t.Fatalf("load command returned %T, want hostnamesLoadedMsg", reopenedMsg)
	}
	updated, finishCmd = reopenedModel.Update(reopened)
	if finishCmd != nil {
		t.Fatal("handling startup load unexpectedly started another command")
	}
	reopenedModel, ok = updated.(Model)
	if !ok {
		t.Fatalf("Update returned unexpected model type %T", updated)
	}
	view := plainText(reopenedModel.View().Content)
	if !strings.Contains(view, "Hostnames guardados") || !strings.Contains(view, `• "example.com"`) {
		t.Fatalf("hostname added in the previous session is not visible after restart:\n%s", view)
	}
}
