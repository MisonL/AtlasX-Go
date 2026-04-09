package main

import (
	"errors"
	"fmt"

	"atlasx/internal/tabs"
)

func runTabsAuthMode(client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs auth-mode")
	}

	view, err := client.AuthMode(args[0])
	if err != nil {
		var captureErr *tabs.CaptureError
		if errors.As(err, &captureErr) {
			printPageContext(captureErr.Context)
		}
		return err
	}

	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s host=%s path=%s mode=%s inferred=%t reason=%s login_prompt_present=%t workspace_signal_present=%t cookie_count=%d local_storage_count=%d session_storage_count=%d\n",
		view.ID,
		view.Title,
		view.URL,
		view.CapturedAt,
		view.Host,
		view.Path,
		view.Mode,
		view.Inferred,
		view.Reason,
		view.LoginPromptPresent,
		view.WorkspaceSignalPresent,
		view.CookieCount,
		view.LocalStorageCount,
		view.SessionStorageCount,
	)
	for index, name := range view.CookieNames {
		fmt.Printf("cookie_names[%d]=%s\n", index, name)
	}
	for index, key := range view.LocalStorageKeys {
		fmt.Printf("local_storage_keys[%d]=%s\n", index, key)
	}
	for index, key := range view.SessionStorageKeys {
		fmt.Printf("session_storage_keys[%d]=%s\n", index, key)
	}
	return nil
}
