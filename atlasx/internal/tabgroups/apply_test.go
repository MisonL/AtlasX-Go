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

func TestSuggestWindowReturnsStructuredGroups(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
					{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
				},
			},
		},
	}

	result, err := SuggestWindow(client, 11)
	if err != nil {
		t.Fatalf("suggest window failed: %v", err)
	}
	if result.SourceWindowID != 11 || result.Returned != 1 || len(result.Groups) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Groups[0].ID != "host:chatgpt.com" {
		t.Fatalf("unexpected groups: %+v", result.Groups)
	}
}

func TestSuggestWindowRejectsUnknownWindow(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
		},
	}

	if _, err := SuggestWindow(client, 11); err == nil {
		t.Fatal("expected suggest window to fail")
	} else if !strings.Contains(err.Error(), "window 11 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyWindowGroupToNewWindowMovesScopedGroup(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
					{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
		},
		moveToNewByID: map[string]tabs.WindowMoveToNewResult{
			"tab-1": {
				SourceWindowID: 11,
				SourceTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
		},
		moveToWindow: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    11,
				TargetWindowID:    21,
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

	result, err := ApplyWindowGroupToNewWindow(client, 11, "host:chatgpt.com")
	if err != nil {
		t.Fatalf("apply window group to new window failed: %v", err)
	}
	if result.SourceWindowID != 11 || result.GroupID != "host:chatgpt.com" || result.WindowID != 21 || result.Returned != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestApplyWindowGroupToNewWindowRejectsUnknownGroup(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
				},
			},
		},
	}

	if _, err := ApplyWindowGroupToNewWindow(client, 11, "host:chatgpt.com"); err == nil {
		t.Fatal("expected apply window group to new window to fail")
	} else if !strings.Contains(err.Error(), "group host:chatgpt.com not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyWindowGroupToWindowMovesScopedGroupIntoTargetWindow(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
					{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "dst-1", Type: "page", Title: "Workspace", URL: "https://workspace.example.com"},
				},
			},
		},
		moveToWindow: map[string]tabs.WindowMoveResult{
			"tab-1": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-1",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-2": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
		},
	}

	result, err := ApplyWindowGroupToWindow(client, 11, "host:chatgpt.com", 21)
	if err != nil {
		t.Fatalf("apply window group to window failed: %v", err)
	}
	if result.SourceWindowID != 11 || result.TargetWindowID != 21 || result.GroupID != "host:chatgpt.com" || result.WindowID != 21 || result.Returned != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.MovedTargets) != 2 || len(result.AlignedTargets) != 0 {
		t.Fatalf("unexpected targets: %+v", result)
	}
}

func TestApplyWindowGroupToWindowRejectsSameWindowID(t *testing.T) {
	client := &stubOrganizerClient{}

	if _, err := ApplyWindowGroupToWindow(client, 11, "host:chatgpt.com", 11); err == nil {
		t.Fatal("expected apply window group to window to fail")
	} else if !strings.Contains(err.Error(), "source and target window ids must differ") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyWindowGroupToWindowRejectsUnknownTargetWindow(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
		},
	}

	if _, err := ApplyWindowGroupToWindow(client, 11, "host:chatgpt.com", 21); err == nil {
		t.Fatal("expected apply window group to window to fail")
	} else if !strings.Contains(err.Error(), "window 21 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyAllToWindowMovesAllSuggestedGroupsIntoTargetWindow(t *testing.T) {
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
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
			{
				WindowID: 7,
				Targets: []tabs.Target{
					{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
					{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
				},
			},
		},
		moveToWindow: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    9,
				TargetWindowID:    11,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
			"tab-3": {
				SourceWindowID:    7,
				TargetWindowID:    11,
				SourceTargetID:    "tab-3",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-3",
					Type:  "page",
					Title: "Build Log - A",
					URL:   "about:blank",
				},
			},
			"tab-4": {
				SourceWindowID:    7,
				TargetWindowID:    11,
				SourceTargetID:    "tab-4",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-4",
					Type:  "page",
					Title: "Build Log - B",
					URL:   "about:blank",
				},
			},
		},
	}

	result, err := ApplyAllToWindow(client, 11)
	if err != nil {
		t.Fatalf("apply all to window failed: %v", err)
	}
	if result.Returned != 2 || len(result.Groups) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Groups[0].GroupID != "title:build log" || result.Groups[1].GroupID != "host:chatgpt.com" {
		t.Fatalf("unexpected groups: %+v", result.Groups)
	}
	if len(result.Groups[1].AlignedTargets) != 1 || len(result.Groups[1].MovedTargets) != 1 {
		t.Fatalf("unexpected host group targets: %+v", result.Groups[1])
	}
}

func TestApplyAllToWindowReturnsEmptyWithoutGroups(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
				},
			},
		},
	}

	result, err := ApplyAllToWindow(client, 11)
	if err != nil {
		t.Fatalf("apply all to window failed: %v", err)
	}
	if result.Returned != 0 || len(result.Groups) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestApplyAllToWindowRejectsUnknownWindow(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
		},
	}

	if _, err := ApplyAllToWindow(client, 11); err == nil {
		t.Fatal("expected apply all to window to fail")
	} else if !strings.Contains(err.Error(), "window 11 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyWindowToNewWindowsMovesSuggestedGroupsFromSourceWindow(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
					{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
					{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
					{ID: "tab-5", Type: "page", Title: "Solo", URL: "https://example.com/solo"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 22,
				Targets: []tabs.Target{
					{ID: "new-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
				},
			},
		},
		moveToNewByID: map[string]tabs.WindowMoveToNewResult{
			"tab-1": {
				SourceWindowID: 11,
				SourceTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-3": {
				SourceWindowID: 11,
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
				SourceWindowID:    11,
				TargetWindowID:    21,
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
				SourceWindowID:    11,
				TargetWindowID:    22,
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

	result, err := ApplyWindowToNewWindows(client, 11)
	if err != nil {
		t.Fatalf("apply window to new windows failed: %v", err)
	}
	if result.SourceWindowID != 11 || result.Returned != 2 || len(result.Groups) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Groups[0].GroupID != "title:build log" || result.Groups[1].GroupID != "host:chatgpt.com" {
		t.Fatalf("unexpected groups: %+v", result.Groups)
	}
}

func TestApplyWindowToNewWindowsReturnsEmptyWithoutGroups(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
				},
			},
		},
	}

	result, err := ApplyWindowToNewWindows(client, 11)
	if err != nil {
		t.Fatalf("apply window to new windows failed: %v", err)
	}
	if result.SourceWindowID != 11 || result.Returned != 0 || len(result.Groups) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestApplyWindowToNewWindowsRejectsUnknownWindow(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
		},
	}

	if _, err := ApplyWindowToNewWindows(client, 11); err == nil {
		t.Fatal("expected apply window to new windows to fail")
	} else if !strings.Contains(err.Error(), "window 11 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyWindowToWindowMovesSuggestedGroupsIntoTargetWindow(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
					{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
					{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
					{ID: "tab-5", Type: "page", Title: "Solo", URL: "https://example.com/solo"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "dst-1", Type: "page", Title: "Workspace", URL: "https://workspace.example.com"},
				},
			},
		},
		moveToWindow: map[string]tabs.WindowMoveResult{
			"tab-1": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-1",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-2": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
			"tab-3": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-3",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-3",
					Type:  "page",
					Title: "Build Log - A",
					URL:   "about:blank",
				},
			},
			"tab-4": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-4",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-4",
					Type:  "page",
					Title: "Build Log - B",
					URL:   "about:blank",
				},
			},
		},
	}

	result, err := ApplyWindowToWindow(client, 11, 21)
	if err != nil {
		t.Fatalf("apply window to window failed: %v", err)
	}
	if result.SourceWindowID != 11 || result.TargetWindowID != 21 || result.Returned != 2 || len(result.Groups) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestApplyWindowToWindowReturnsEmptyWithoutGroups(t *testing.T) {
	client := &stubOrganizerClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "dst-1", Type: "page", Title: "Workspace", URL: "https://workspace.example.com"},
				},
			},
		},
	}

	result, err := ApplyWindowToWindow(client, 11, 21)
	if err != nil {
		t.Fatalf("apply window to window failed: %v", err)
	}
	if result.SourceWindowID != 11 || result.TargetWindowID != 21 || result.Returned != 0 || len(result.Groups) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestApplyWindowToWindowRejectsSameWindowID(t *testing.T) {
	client := &stubOrganizerClient{}

	if _, err := ApplyWindowToWindow(client, 11, 11); err == nil {
		t.Fatal("expected apply window to window to fail")
	} else if !strings.Contains(err.Error(), "source and target window ids must differ") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyGroupToWindowMovesTargetsAndKeepsAlignedOnes(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
		},
		moveToWindow: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    9,
				TargetWindowID:    11,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
		},
	}

	result, err := ApplyGroupToWindow(client, "host:chatgpt.com", 11)
	if err != nil {
		t.Fatalf("apply group to window failed: %v", err)
	}
	if result.WindowID != 11 || result.Returned != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.AlignedTargets) != 1 || result.AlignedTargets[0].SourceTargetID != "tab-1" {
		t.Fatalf("unexpected aligned targets: %+v", result.AlignedTargets)
	}
	if len(result.MovedTargets) != 1 || result.MovedTargets[0].SourceTargetID != "tab-2" {
		t.Fatalf("unexpected moved targets: %+v", result.MovedTargets)
	}
}

func TestApplyGroupToWindowRejectsUnknownWindow(t *testing.T) {
	client := &stubOrganizerClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
		},
	}

	if _, err := ApplyGroupToWindow(client, "host:chatgpt.com", 11); err == nil {
		t.Fatal("expected apply group to window to fail")
	} else if !strings.Contains(err.Error(), "window 11 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type errString string

func (e errString) Error() string {
	return string(e)
}
