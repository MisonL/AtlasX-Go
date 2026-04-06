package managedruntime

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestLoadCatalogFromFileAndResolveInstallPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	catalogPath := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(catalogPath, []byte(`{
  "entries": [
    {
      "platform": "darwin-amd64",
      "channel": "stable",
      "version": "123.0.0",
      "url": "https://example.com/chromium-123.zip",
      "sha256": "deadbeef",
      "bundle_name": "Chromium.app"
    }
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write catalog failed: %v", err)
	}

	catalog, err := LoadCatalog(catalogPath, nil)
	if err != nil {
		t.Fatalf("load catalog failed: %v", err)
	}

	plan, err := ResolveInstallPlanFromCatalog(catalog, ResolveCatalogOptions{
		Version:          "123.0.0",
		Channel:          "stable",
		Platform:         DefaultCatalogPlatform,
		ArchivePath:      DefaultInstallArchivePath(paths, "123.0.0"),
		StagedBundlePath: DefaultStagedBundlePath(paths),
	})
	if err != nil {
		t.Fatalf("resolve install plan failed: %v", err)
	}
	if plan.BundleName != "Chromium.app" || plan.SourceURL != "https://example.com/chromium-123.zip" {
		t.Fatalf("unexpected resolved plan: %+v", plan)
	}
}

func TestLoadCatalogFromHTTPS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "entries": [
    {
      "platform": "darwin-amd64",
      "channel": "stable",
      "version": "123.0.0",
      "url": "https://example.com/chromium-123.zip",
      "sha256": "deadbeef",
      "bundle_name": "Chromium.app"
    }
  ]
}`))
	}))
	defer server.Close()

	catalog, err := LoadCatalog(server.URL, server.Client())
	if err != nil {
		t.Fatalf("load remote catalog failed: %v", err)
	}
	if len(catalog.Entries) != 1 {
		t.Fatalf("unexpected remote catalog: %+v", catalog)
	}
}

func TestResolveInstallPlanRejectsMissingVersion(t *testing.T) {
	_, err := ResolveInstallPlanFromCatalog(Catalog{}, ResolveCatalogOptions{
		Channel:          "stable",
		Platform:         DefaultCatalogPlatform,
		ArchivePath:      "/tmp/chromium.zip",
		StagedBundlePath: "/tmp/Chromium.app",
	})
	if !errors.Is(err, ErrCatalogVersionRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveInstallPlanRejectsMissingCatalogEntry(t *testing.T) {
	_, err := ResolveInstallPlanFromCatalog(Catalog{
		Entries: []CatalogEntry{{
			Platform:   DefaultCatalogPlatform,
			Channel:    "beta",
			Version:    "124.0.0",
			SourceURL:  "https://example.com/chromium-124.zip",
			SHA256:     "deadbeef",
			BundleName: "Chromium.app",
		}},
	}, ResolveCatalogOptions{
		Version:          "123.0.0",
		Channel:          "stable",
		Platform:         DefaultCatalogPlatform,
		ArchivePath:      "/tmp/chromium.zip",
		StagedBundlePath: "/tmp/Chromium.app",
	})
	if !errors.Is(err, ErrCatalogEntryNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}
