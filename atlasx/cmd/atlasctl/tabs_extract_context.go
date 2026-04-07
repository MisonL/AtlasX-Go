package main

import (
	"errors"
	"fmt"

	"atlasx/internal/tabs"
)

func runTabsExtractContext(client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs extract-context")
	}

	context, err := client.CaptureSemanticContext(args[0])
	if err != nil {
		printSemanticContext(context)
		return err
	}
	printSemanticContext(context)
	return nil
}

func printSemanticContext(context tabs.SemanticContext) {
	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s returned=%d headings_returned=%d links_returned=%d forms_returned=%d capture_error=%q\n",
		context.ID,
		context.Title,
		context.URL,
		context.CapturedAt,
		context.Returned,
		context.HeadingsReturned,
		context.LinksReturned,
		context.FormsReturned,
		context.CaptureError,
	)
	for index, heading := range context.Headings {
		fmt.Printf("heading_index=%d level=%d text=%q\n", index, heading.Level, heading.Text)
	}
	for index, link := range context.Links {
		fmt.Printf("link_index=%d text=%q url=%s\n", index, link.Text, link.URL)
	}
	for index, form := range context.Forms {
		fmt.Printf("form_index=%d action=%s method=%s input_count=%d\n", index, form.Action, form.Method, form.InputCount)
	}
}
