package main

import (
	"reflect"
	"testing"
)

func TestPrepareCodexArgs_ModelInjection(t *testing.T) {
	env := Environment{
		Name:   "dev",
		URL:    "https://api.openai.com/v1",
		APIKey: "sk-test",
		Model:  "gpt-5",
	}

	// Case 1: user provides no -m/--model, should inject
	in := []string{"exec", "--fast"}
	out := prepareCodexArgs(env, in)
	expectedPrefix := []string{"-m", "gpt-5"}
	if len(out) < 2 || !reflect.DeepEqual(out[:2], expectedPrefix) {
		t.Fatalf("expected prefix %v, got %v", expectedPrefix, out)
	}

	// Case 2: user provides -m, should not override
	in2 := []string{"-m", "o4-mini", "proto"}
	out2 := prepareCodexArgs(env, in2)
	if !reflect.DeepEqual(out2, in2) {
		t.Fatalf("expected no change when -m present, got %v", out2)
	}
}

func TestApplyAutoFlags(t *testing.T) {
	args := []string{"proto"}
	result := applyAutoFlags(args)
	if len(result) != len(args)+4 {
		t.Fatalf("expected 4 flags added, got %d: %v", len(result)-len(args), result)
	}
	expected := []string{"-a", "never", "--sandbox", "workspace-write"}
	for i := 0; i < 4; i++ {
		if result[i] != expected[i] {
			t.Fatalf("expected %v at %d, got %v", expected, i, result)
		}
	}
}
