package main

import "testing"

func TestValidateListenAddressAllowsLoopbackHosts(t *testing.T) {
	testCases := []string{
		"127.0.0.1:17537",
		"localhost:17537",
		"[::1]:17537",
	}

	for _, testCase := range testCases {
		if err := validateListenAddress(testCase, false); err != nil {
			t.Fatalf("expected listen address %s to pass: %v", testCase, err)
		}
	}
}

func TestValidateListenAddressRejectsNonLoopbackByDefault(t *testing.T) {
	testCases := []string{
		"0.0.0.0:17537",
		"192.168.1.10:17537",
		":17537",
	}

	for _, testCase := range testCases {
		if err := validateListenAddress(testCase, false); err == nil {
			t.Fatalf("expected listen address %s to fail", testCase)
		}
	}
}

func TestValidateListenAddressAllowsRemoteWhenExplicitlyEnabled(t *testing.T) {
	if err := validateListenAddress("0.0.0.0:17537", true); err != nil {
		t.Fatalf("expected remote listen address to pass with override: %v", err)
	}
}
