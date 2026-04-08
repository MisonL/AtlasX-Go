package defaultbrowser

import (
	"errors"
	"strings"
	"testing"
)

func TestReadStatusParsesHTTPAndHTTPSHandlers(t *testing.T) {
	previous := readLaunchServicesJSON
	readLaunchServicesJSON = func() ([]byte, error) {
		return []byte(`{"LSHandlers":[{"LSHandlerURLScheme":"http","LSHandlerRoleAll":"org.mozilla.firefox"},{"LSHandlerURLScheme":"https","LSHandlerRoleAll":"org.mozilla.firefox"}]}`), nil
	}
	t.Cleanup(func() {
		readLaunchServicesJSON = previous
	})

	status, err := ReadStatus()
	if err != nil {
		t.Fatalf("read status failed: %v", err)
	}

	if status.HTTPBundleID != "org.mozilla.firefox" {
		t.Fatalf("unexpected http bundle id: %s", status.HTTPBundleID)
	}
	if status.HTTPSBundleID != "org.mozilla.firefox" {
		t.Fatalf("unexpected https bundle id: %s", status.HTTPSBundleID)
	}
	if status.HTTPRole != "all" || status.HTTPSRole != "all" {
		t.Fatalf("unexpected roles: %+v", status)
	}
	if !status.HTTPKnown || !status.HTTPSKnown {
		t.Fatalf("expected schemes to be known: %+v", status)
	}
	if !status.Consistent {
		t.Fatalf("expected consistent status: %+v", status)
	}
}

func TestReadStatusMarksMissingSchemeUnknown(t *testing.T) {
	previous := readLaunchServicesJSON
	readLaunchServicesJSON = func() ([]byte, error) {
		return []byte(`{"LSHandlers":[{"LSHandlerURLScheme":"http","LSHandlerRoleViewer":"com.apple.Safari"}]}`), nil
	}
	t.Cleanup(func() {
		readLaunchServicesJSON = previous
	})

	status, err := ReadStatus()
	if err != nil {
		t.Fatalf("read status failed: %v", err)
	}

	if status.HTTPBundleID != "com.apple.Safari" || status.HTTPRole != "viewer" || !status.HTTPKnown {
		t.Fatalf("unexpected http status: %+v", status)
	}
	if status.HTTPSBundleID != "unknown" || status.HTTPSRole != "unknown" || status.HTTPSKnown {
		t.Fatalf("expected unknown https status: %+v", status)
	}
	if status.Consistent {
		t.Fatalf("expected inconsistent status when https is missing: %+v", status)
	}
}

func TestReadStatusReturnsCommandError(t *testing.T) {
	previous := readLaunchServicesJSON
	readLaunchServicesJSON = func() ([]byte, error) {
		return nil, errors.New("defaults failed")
	}
	t.Cleanup(func() {
		readLaunchServicesJSON = previous
	})

	_, err := ReadStatus()
	if err == nil {
		t.Fatal("expected read status to fail")
	}
	if !strings.Contains(err.Error(), "defaults failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadStatusRejectsMalformedJSON(t *testing.T) {
	previous := readLaunchServicesJSON
	readLaunchServicesJSON = func() ([]byte, error) {
		return []byte(`not-json`), nil
	}
	t.Cleanup(func() {
		readLaunchServicesJSON = previous
	})

	_, err := ReadStatus()
	if err == nil {
		t.Fatal("expected malformed json to fail")
	}
	if !strings.Contains(err.Error(), "decode launchservices handlers") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetBundleIDRewritesHTTPAndHTTPSHandlers(t *testing.T) {
	var stored []byte

	previousRead := readLaunchServicesJSON
	previousWrite := writeLaunchServicesJSON
	readLaunchServicesJSON = func() ([]byte, error) {
		if stored != nil {
			return stored, nil
		}
		return []byte(`{"LSHandlers":[{"LSHandlerURLScheme":"mailto","LSHandlerRoleAll":"com.apple.mail"},{"LSHandlerURLScheme":"http","LSHandlerRoleViewer":"org.mozilla.firefox"},{"LSHandlerURLScheme":"https","LSHandlerRoleAll":"org.mozilla.firefox"}]}`), nil
	}
	writeLaunchServicesJSON = func(payload []byte) error {
		stored = append([]byte(nil), payload...)
		return nil
	}
	t.Cleanup(func() {
		readLaunchServicesJSON = previousRead
		writeLaunchServicesJSON = previousWrite
	})

	status, err := SetBundleID("com.openai.atlasx")
	if err != nil {
		t.Fatalf("set bundle id failed: %v", err)
	}
	if status.HTTPBundleID != "com.openai.atlasx" || status.HTTPSBundleID != "com.openai.atlasx" || !status.Consistent {
		t.Fatalf("unexpected status: %+v", status)
	}

	handlers, err := parseHandlers(stored)
	if err != nil {
		t.Fatalf("parse stored handlers failed: %v", err)
	}
	if len(handlers.LSHandlers) != 3 {
		t.Fatalf("unexpected handlers: %+v", handlers.LSHandlers)
	}
	if handlers.LSHandlers[0].URLScheme != "mailto" {
		t.Fatalf("expected unrelated scheme to be preserved: %+v", handlers.LSHandlers)
	}
}

func TestSetBundleIDRejectsBlankBundleID(t *testing.T) {
	_, err := SetBundleID("  ")
	if err == nil {
		t.Fatal("expected blank bundle id to fail")
	}
	if !strings.Contains(err.Error(), "bundle_id is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetBundleIDReturnsWriteError(t *testing.T) {
	previousRead := readLaunchServicesJSON
	previousWrite := writeLaunchServicesJSON
	readLaunchServicesJSON = func() ([]byte, error) {
		return []byte(`{"LSHandlers":[]}`), nil
	}
	writeLaunchServicesJSON = func(payload []byte) error {
		return errors.New("import failed")
	}
	t.Cleanup(func() {
		readLaunchServicesJSON = previousRead
		writeLaunchServicesJSON = previousWrite
	})

	_, err := SetBundleID("com.openai.atlasx")
	if err == nil {
		t.Fatal("expected write error")
	}
	if !strings.Contains(err.Error(), "import failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetBundleIDReturnsVerificationError(t *testing.T) {
	previousRead := readLaunchServicesJSON
	previousWrite := writeLaunchServicesJSON
	callCount := 0
	readLaunchServicesJSON = func() ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(`{"LSHandlers":[{"LSHandlerURLScheme":"http","LSHandlerRoleAll":"org.mozilla.firefox"}]}`), nil
		}
		return []byte(`{"LSHandlers":[{"LSHandlerURLScheme":"http","LSHandlerRoleAll":"org.mozilla.firefox"},{"LSHandlerURLScheme":"https","LSHandlerRoleAll":"org.mozilla.firefox"}]}`), nil
	}
	writeLaunchServicesJSON = func(payload []byte) error {
		return nil
	}
	t.Cleanup(func() {
		readLaunchServicesJSON = previousRead
		writeLaunchServicesJSON = previousWrite
	})

	_, err := SetBundleID("com.openai.atlasx")
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !strings.Contains(err.Error(), "did not apply bundle_id") {
		t.Fatalf("unexpected error: %v", err)
	}
}
