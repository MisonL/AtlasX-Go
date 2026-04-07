package tabs

import (
	"net/url"
	"strings"
)

type DuplicateCloseGroup struct {
	URL             string   `json:"url"`
	KeptTargetID    string   `json:"kept_target_id"`
	ClosedTargetIDs []string `json:"closed_target_ids"`
	Returned        int      `json:"returned"`
}

type CloseDuplicatesResult struct {
	Returned      int                   `json:"returned"`
	Groups        []DuplicateCloseGroup `json:"groups"`
	ClosedTargets []string              `json:"closed_targets"`
}

func (c Client) CloseDuplicates() (CloseDuplicatesResult, error) {
	targets, err := c.List()
	if err != nil {
		return CloseDuplicatesResult{}, err
	}

	tracked, order := duplicateClosePlan(PageTargets(targets))
	result := CloseDuplicatesResult{
		Groups:        make([]DuplicateCloseGroup, 0),
		ClosedTargets: make([]string, 0),
	}
	for _, key := range order {
		group := tracked[key]
		for _, targetID := range group.ClosedTargetIDs {
			if err := c.Close(targetID); err != nil {
				return CloseDuplicatesResult{}, err
			}
			result.ClosedTargets = append(result.ClosedTargets, targetID)
			result.Returned++
		}
		if group.Returned > 0 {
			result.Groups = append(result.Groups, *group)
		}
	}
	return result, nil
}

func duplicateClosePlan(targets []Target) (map[string]*DuplicateCloseGroup, []string) {
	tracked := make(map[string]*DuplicateCloseGroup)
	order := make([]string, 0)
	for _, target := range targets {
		key, ok := duplicateTargetKey(target.URL)
		if !ok {
			continue
		}
		group, exists := tracked[key]
		if !exists {
			tracked[key] = &DuplicateCloseGroup{
				URL:             key,
				KeptTargetID:    target.ID,
				ClosedTargetIDs: make([]string, 0),
			}
			order = append(order, key)
			continue
		}
		group.ClosedTargetIDs = append(group.ClosedTargetIDs, target.ID)
		group.Returned = len(group.ClosedTargetIDs)
	}
	return tracked, order
}

func duplicateTargetKey(raw string) (string, bool) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", false
	}
	host := duplicateTargetHost(parsed)
	if host == "" {
		return "", false
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	key := scheme + "://" + host + path
	if parsed.RawQuery != "" {
		key += "?" + parsed.RawQuery
	}
	return key, true
}

func duplicateTargetHost(parsed *url.URL) string {
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return ""
	}
	port := parsed.Port()
	if port == "" || (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
		return host
	}
	return host + ":" + port
}
