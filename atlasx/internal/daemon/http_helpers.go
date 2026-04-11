package daemon

import (
	"fmt"
	"net/http"
	"strings"
)

func requireMethods(next http.HandlerFunc, methods ...string) http.HandlerFunc {
	allowed := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		allowed[method] = struct{}{}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[r.Method]; !ok {
			writeMethodNotAllowed(w, r.Method, methods...)
			return
		}
		next(w, r)
	}
}

func writeMethodNotAllowed(w http.ResponseWriter, got string, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", got))
}

func writeTabsClientUnavailable(w http.ResponseWriter, err error) {
	writeError(w, http.StatusServiceUnavailable, err)
}

func writeSidebarTabsClientUnavailable(w http.ResponseWriter, traceID string, err error) {
	writeSidebarAskError(w, http.StatusServiceUnavailable, traceID, err)
}

func truncateTextSnippet(text string, limit int) string {
	runes := []rune(text)
	if limit <= 0 || len(runes) <= limit {
		return text
	}
	return string(runes[:limit])
}
