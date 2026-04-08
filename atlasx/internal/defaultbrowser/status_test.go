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
