package tabgroups

import (
	"net/url"
	"sort"
	"strings"

	"atlasx/internal/tabs"
)

type Group struct {
	ID       string        `json:"id"`
	Label    string        `json:"label"`
	Reason   string        `json:"reason"`
	Targets  []tabs.Target `json:"targets"`
	Returned int           `json:"returned"`
}

func Suggest(targets []tabs.Target) []Group {
	pages := tabs.PageTargets(targets)
	byKey := make(map[string][]tabs.Target)
	order := make([]string, 0)

	for _, target := range pages {
		key, label, ok := groupingKey(target)
		if !ok {
			continue
		}
		if _, exists := byKey[key]; !exists {
			order = append(order, key+"\x00"+label)
		}
		byKey[key] = append(byKey[key], target)
	}

	groups := make([]Group, 0)
	for _, entry := range order {
		parts := strings.SplitN(entry, "\x00", 2)
		key := parts[0]
		label := parts[1]
		groupTargets := byKey[key]
		if len(groupTargets) < 2 {
			continue
		}
		sort.SliceStable(groupTargets, func(i, j int) bool {
			if groupTargets[i].Title == groupTargets[j].Title {
				return groupTargets[i].ID < groupTargets[j].ID
			}
			return groupTargets[i].Title < groupTargets[j].Title
		})
		groups = append(groups, Group{
			ID:       key,
			Label:    label,
			Reason:   buildReason(key),
			Targets:  append([]tabs.Target(nil), groupTargets...),
			Returned: len(groupTargets),
		})
	}

	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Returned == groups[j].Returned {
			return groups[i].Label < groups[j].Label
		}
		return groups[i].Returned > groups[j].Returned
	})
	return groups
}

func groupingKey(target tabs.Target) (string, string, bool) {
	host := normalizedHost(target.URL)
	if host != "" {
		return "host:" + host, host, true
	}

	prefix := titlePrefix(target.Title)
	if prefix != "" {
		return "title:" + prefix, prefix, true
	}
	return "", "", false
}

func normalizedHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := strings.TrimSpace(strings.ToLower(parsed.Hostname()))
	if host == "" {
		return ""
	}
	return host
}

func titlePrefix(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return ""
	}
	separators := []string{" - ", " | ", " — ", " · ", ":"}
	for _, separator := range separators {
		if parts := strings.SplitN(trimmed, separator, 2); len(parts) == 2 {
			prefix := strings.TrimSpace(parts[0])
			if len(prefix) >= 4 {
				return strings.ToLower(prefix)
			}
		}
	}
	if len(trimmed) >= 8 {
		return strings.ToLower(trimmed)
	}
	return ""
}

func buildReason(key string) string {
	switch {
	case strings.HasPrefix(key, "host:"):
		return "These tabs share the same host and are likely part of the same browsing thread."
	case strings.HasPrefix(key, "title:"):
		return "These tabs share the same title prefix and likely belong to the same task."
	default:
		return "These tabs appear related."
	}
}
