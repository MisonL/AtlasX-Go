package main

import (
	"errors"
	"fmt"
	"strings"
)

func runTabsSearch(client commandTabsClient, args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return errors.New("missing query for tabs search")
	}

	targets, err := client.Search(query)
	if err != nil {
		return err
	}

	fmt.Printf("returned=%d\n", len(targets))
	for index, target := range targets {
		fmt.Printf("index=%d id=%s type=%s title=%q url=%s\n", index, target.ID, target.Type, target.Title, target.URL)
	}
	return nil
}
