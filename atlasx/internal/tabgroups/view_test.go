package tabgroups

import (
	"errors"
	"testing"

	"atlasx/internal/tabs"
)

type stubGroupsClient struct {
	windows []tabs.WindowSummary
	err     error
}

func (s stubGroupsClient) Windows() ([]tabs.WindowSummary, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.windows, nil
}

func TestBuildViewIncludesWindowMetadata(t *testing.T) {
	result := BuildView([]tabs.WindowSummary{
		{
			WindowID: 11,
			Targets: []tabs.Target{
				{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
			},
		},
		{
			WindowID: 22,
			Targets: []tabs.Target{
				{ID: "tab-3", Type: "page", Title: "Atlas C", URL: "https://chatgpt.com/atlas/c"},
				{ID: "tab-4", Type: "page", Title: "Other", URL: "https://example.com/other"},
			},
		},
	})

	if !result.Inferred || result.Returned != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	group := result.Groups[0]
	if !group.Inferred || group.ID != "host:chatgpt.com" || group.Returned != 3 || group.WindowReturned != 2 {
		t.Fatalf("unexpected group: %+v", group)
	}
	if len(group.WindowIDs) != 2 || group.WindowIDs[0] != 11 || group.WindowIDs[1] != 22 {
		t.Fatalf("unexpected window ids: %+v", group.WindowIDs)
	}
	if len(group.Windows) != 2 || group.Windows[0].Returned != 2 || group.Windows[1].Returned != 1 {
		t.Fatalf("unexpected group windows: %+v", group.Windows)
	}
}

func TestInspectPropagatesWindowsError(t *testing.T) {
	expected := errors.New("windows unavailable")

	_, err := Inspect(stubGroupsClient{err: expected})
	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
