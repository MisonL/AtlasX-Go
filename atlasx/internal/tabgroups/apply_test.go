package tabgroups

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

type stubOrganizerClient struct {
	targets         []tabs.Target
	listErr         error
	windows         []tabs.WindowSummary
	windowsErr      error
	moveToWindow    map[string]tabs.WindowMoveResult
	moveToWindowErr error
	moveToNew       tabs.WindowMoveToNewResult
	moveToNewByID   map[string]tabs.WindowMoveToNewResult
	moveToNewErr    error
	moveWindowID    int
	moveCalls       []string
}

func (s *stubOrganizerClient) List() ([]tabs.Target, error) {
	return s.targets, s.listErr
}

func (s *stubOrganizerClient) Windows() ([]tabs.WindowSummary, error) {
	return s.windows, s.windowsErr
}

func (s *stubOrganizerClient) MoveToWindow(targetID string, targetWindowID int) (tabs.WindowMoveResult, error) {
	s.moveCalls = append(s.moveCalls, targetID)
	s.moveWindowID = targetWindowID
	if s.moveToWindowErr != nil {
		return tabs.WindowMoveResult{}, s.moveToWindowErr
	}
	return s.moveToWindow[targetID], nil
}

func (s *stubOrganizerClient) MoveToNewWindow(targetID string) (tabs.WindowMoveToNewResult, error) {
	s.moveCalls = append(s.moveCalls, "new:"+targetID)
	if s.moveToNewErr != nil {
		return tabs.WindowMoveToNewResult{}, s.moveToNewErr
	}
	if result, ok := s.moveToNewByID[targetID]; ok {
		return result, nil
	}
	return s.moveToNew, nil
}

func TestApplyToNewWindowMovesSuggestedGroup(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
			{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
		},
		moveToNew: tabs.WindowMoveToNewResult{
			SourceWindowID: 9,
			SourceTargetID: "tab-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "Atlas A",
				URL:   "https://chatgpt.com/atlas/a",
			},
		},
		moveToWindow: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    9,
				TargetWindowID:    11,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "new-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
		},
	}

	result, err := ApplyToNewWindow(client, "host:chatgpt.com")
	if err != nil {
		t.Fatalf("apply to new window failed: %v", err)
	}
	if result.GroupID != "host:chatgpt.com" || result.WindowID != 11 || result.Returned != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.MovedTargets) != 2 || result.MovedTargets[0].SourceTargetID != "tab-1" || result.MovedTargets[1].SourceTargetID != "tab-2" {
		t.Fatalf("unexpected moved targets: %+v", result.MovedTargets)
	}
	if client.moveWindowID != 11 {
		t.Fatalf("unexpected target window id: %d", client.moveWindowID)
	}
}

func TestApplyToNewWindowRejectsUnknownGroup(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
		},
	}

	if _, err := ApplyToNewWindow(client, "host:chatgpt.com"); err == nil {
		t.Fatal("expected apply to new window to fail")
	} else if !strings.Contains(err.Error(), "group host:chatgpt.com not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyToNewWindowSurfacesMoveFailure(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
		},
		moveToNew: tabs.WindowMoveToNewResult{
			SourceWindowID: 9,
			SourceTargetID: "tab-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "Atlas A",
				URL:   "https://chatgpt.com/atlas/a",
			},
		},
		moveToWindowErr: errString("move failed"),
	}

	if _, err := ApplyToNewWindow(client, "host:chatgpt.com"); err == nil {
		t.Fatal("expected apply to new window to fail")
	} else if !strings.Contains(err.Error(), "move failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyAllToNewWindowsMovesAllSuggestedGroups(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
			{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
			{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 12,
				Targets: []tabs.Target{
					{ID: "new-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
				},
			},
		},
		moveToNewByID: map[string]tabs.WindowMoveToNewResult{
			"tab-1": {
				SourceWindowID: 9,
				SourceTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-3": {
				SourceWindowID: 9,
				SourceTargetID: "tab-3",
				Target: tabs.Target{
					ID:    "new-3",
					Type:  "page",
					Title: "Build Log - A",
					URL:   "about:blank",
				},
			},
		},
		moveToWindow: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    9,
				TargetWindowID:    11,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "new-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
			"tab-4": {
				SourceWindowID:    9,
				TargetWindowID:    12,
				SourceTargetID:    "tab-4",
				ActivatedTargetID: "new-3",
				Target: tabs.Target{
					ID:    "new-4",
					Type:  "page",
					Title: "Build Log - B",
					URL:   "about:blank",
				},
			},
		},
	}

	result, err := ApplyAllToNewWindows(client)
	if err != nil {
		t.Fatalf("apply all to new windows failed: %v", err)
	}
	if result.Returned != 2 || len(result.Groups) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Groups[0].GroupID != "title:build log" || result.Groups[1].GroupID != "host:chatgpt.com" {
		t.Fatalf("unexpected groups: %+v", result.Groups)
	}
}

func TestApplyAllToNewWindowsReturnsEmptyWithoutGroups(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
		},
	}

	result, err := ApplyAllToNewWindows(client)
	if err != nil {
		t.Fatalf("apply all to new windows failed: %v", err)
	}
	if result.Returned != 0 || len(result.Groups) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

type errString string

func (e errString) Error() string {
	return string(e)
}
