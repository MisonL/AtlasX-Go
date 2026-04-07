package tabs

import "strings"

func (c Client) Search(query string) ([]Target, error) {
	targets, err := c.List()
	if err != nil {
		return nil, err
	}
	return SearchTargets(targets, query), nil
}

func SearchTargets(targets []Target, query string) []Target {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return []Target{}
	}

	pages := PageTargets(targets)
	matches := make([]Target, 0)
	for _, target := range pages {
		searchable := strings.ToLower(target.Title + "\n" + target.URL)
		if strings.Contains(searchable, normalizedQuery) {
			matches = append(matches, target)
		}
	}
	return matches
}
