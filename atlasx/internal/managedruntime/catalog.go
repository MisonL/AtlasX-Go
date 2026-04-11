package managedruntime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"atlasx/internal/platform/macos"
)

const DefaultCatalogPlatform = "darwin-amd64"

var (
	ErrCatalogSourceRequired   = errors.New("managed runtime catalog source is required")
	ErrCatalogSourceInvalid    = errors.New("managed runtime catalog source must be a local path or https url")
	ErrCatalogVersionRequired  = errors.New("managed runtime catalog resolve version is required")
	ErrCatalogChannelRequired  = errors.New("managed runtime catalog resolve channel is required")
	ErrCatalogEntryNotFound    = errors.New("managed runtime catalog entry was not found")
	ErrCatalogBundleNameAbsent = errors.New("managed runtime catalog bundle_name is required")
)

type Catalog struct {
	Entries []CatalogEntry `json:"entries"`
}

type CatalogEntry struct {
	Platform   string `json:"platform"`
	Channel    string `json:"channel"`
	Version    string `json:"version"`
	SourceURL  string `json:"url"`
	SHA256     string `json:"sha256"`
	BundleName string `json:"bundle_name"`
}

type ResolveCatalogOptions struct {
	Version          string
	Channel          string
	Platform         string
	ArchivePath      string
	StagedBundlePath string
}

func LoadCatalog(source string, client *http.Client) (Catalog, error) {
	if source == "" {
		return Catalog{}, ErrCatalogSourceRequired
	}

	switch {
	case strings.HasPrefix(source, "https://"):
		return loadCatalogFromHTTPS(source, client)
	case strings.HasPrefix(source, "http://"):
		return Catalog{}, ErrCatalogSourceInvalid
	default:
		return loadCatalogFromFile(source)
	}
}

func ResolveInstallPlanFromCatalog(catalog Catalog, opts ResolveCatalogOptions) (InstallPlan, error) {
	if opts.Version == "" {
		return InstallPlan{}, ErrCatalogVersionRequired
	}
	if opts.Channel == "" {
		return InstallPlan{}, ErrCatalogChannelRequired
	}
	if opts.Platform == "" {
		opts.Platform = DefaultCatalogPlatform
	}
	if opts.ArchivePath == "" {
		return InstallPlan{}, ErrInstallPlanArchivePathRequired
	}
	if opts.StagedBundlePath == "" {
		return InstallPlan{}, ErrInstallPlanBundlePathRequired
	}

	entry, err := catalog.resolveEntry(opts.Platform, opts.Channel, opts.Version)
	if err != nil {
		return InstallPlan{}, err
	}

	return NewInstallPlan(InstallPlanOptions{
		Version:          entry.Version,
		Channel:          entry.Channel,
		BundleName:       entry.BundleName,
		SourceURL:        entry.SourceURL,
		ExpectedSHA256:   entry.SHA256,
		ArchivePath:      opts.ArchivePath,
		StagedBundlePath: opts.StagedBundlePath,
	})
}

func DefaultInstallArchivePath(paths macos.Paths, version string) string {
	filename := "chromium-catalog.zip"
	if version != "" {
		filename = fmt.Sprintf("chromium-%s.zip", version)
	}
	return filepath.Join(paths.RuntimeRoot, "downloads", filename)
}

func DefaultStagedBundlePath(paths macos.Paths) string {
	return filepath.Join(paths.RuntimeRoot, stagedBundleName)
}

func loadCatalogFromFile(source string) (Catalog, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return Catalog{}, err
	}
	return decodeCatalog(data)
}

func loadCatalogFromHTTPS(source string, client *http.Client) (Catalog, error) {
	if client == nil {
		client = http.DefaultClient
	}

	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return Catalog{}, ErrCatalogSourceInvalid
	}

	response, err := client.Get(source)
	if err != nil {
		return Catalog{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return Catalog{}, fmt.Errorf("managed runtime catalog returned status %d", response.StatusCode)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return Catalog{}, err
	}
	return decodeCatalog(data)
}

func decodeCatalog(data []byte) (Catalog, error) {
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, err
	}
	for _, entry := range catalog.Entries {
		if entry.BundleName == "" {
			return Catalog{}, ErrCatalogBundleNameAbsent
		}
	}
	return catalog, nil
}

func (c Catalog) resolveEntry(platform string, channel string, version string) (CatalogEntry, error) {
	for _, entry := range c.Entries {
		if entry.Platform == platform && entry.Channel == channel && entry.Version == version {
			return entry, nil
		}
	}
	return CatalogEntry{}, fmt.Errorf(
		"%w: platform=%s channel=%s version=%s",
		ErrCatalogEntryNotFound,
		platform,
		channel,
		version,
	)
}
