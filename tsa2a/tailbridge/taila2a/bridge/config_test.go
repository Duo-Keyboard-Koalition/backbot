package main

import "testing"

func TestGetenvUsesFallback(t *testing.T) {
	t.Setenv("BRIDGE_TEST_ENV", "")
	if got := getenv("BRIDGE_TEST_ENV", "fallback"); got != "fallback" {
		t.Fatalf("getenv() = %q, want fallback", got)
	}

	t.Setenv("BRIDGE_TEST_ENV", "set-value")
	if got := getenv("BRIDGE_TEST_ENV", "fallback"); got != "set-value" {
		t.Fatalf("getenv() = %q, want set-value", got)
	}
}

func TestGetenvInt(t *testing.T) {
	t.Setenv("BRIDGE_TEST_INT", "9001")
	if got := getenvInt("BRIDGE_TEST_INT", 8001); got != 9001 {
		t.Fatalf("getenvInt() = %d, want 9001", got)
	}

	t.Setenv("BRIDGE_TEST_INT", "not-a-number")
	if got := getenvInt("BRIDGE_TEST_INT", 8001); got != 8001 {
		t.Fatalf("getenvInt() invalid fallback = %d, want 8001", got)
	}

	t.Setenv("BRIDGE_TEST_INT", "")
	if got := getenvInt("BRIDGE_TEST_INT", 8001); got != 8001 {
		t.Fatalf("getenvInt() empty fallback = %d, want 8001", got)
	}
}
