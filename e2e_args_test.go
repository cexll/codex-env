package main

import (
	"reflect"
	"testing"
)

// E2E-ish: ensure runAuto will prepend auto flags and model injection when needed
func TestRunAutoArgsAssembly(t *testing.T) {
	env := Environment{
		Name:   "dev",
		URL:    "https://api.openai.com/v1",
		APIKey: "sk-test",
		Model:  "gpt-5",
	}

	// No -m provided by user; expect model injection and auto flags
	user := []string{"proto"}
	injected := prepareCodexArgs(env, user)
	// model injection should add -m gpt-5 at the front
	wantPrefix := []string{"-m", "gpt-5"}
	if len(injected) < 2 || !reflect.DeepEqual(injected[:2], wantPrefix) {
		t.Fatalf("expected model injection prefix %v, got %v", wantPrefix, injected)
	}

	auto := applyAutoFlags(injected)
	wantAutoPrefix := []string{"-a", "never", "--sandbox", "workspace-write"}
	if len(auto) < 6 || !reflect.DeepEqual(auto[:4], wantAutoPrefix) {
		t.Fatalf("expected auto flags prefix %v, got %v", wantAutoPrefix, auto)
	}
}
