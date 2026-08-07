package blackbelt

import (
	"encoding/json"
	"strconv"
	"strings"
)

func object(v any) map[string]any { m, _ := v.(map[string]any); return m }
func array(v any) []any           { a, _ := v.([]any); return a }
func objects(v []any) []map[string]any {
	out := []map[string]any{}
	for _, x := range v {
		if o := object(x); o != nil {
			out = append(out, o)
		}
	}
	return out
}
func stringValue(v any) string { s, _ := v.(string); return s }
func intValue(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	}
	return 0
}
func int64Value(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case json.Number:
		n, _ := x.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	}
	return 0
}
func boolValue(v any) bool { b, _ := v.(bool); return b }
func nonempty(s string) []string {
	var out []string
	for _, v := range strings.Split(s, "\n") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
func keys[T comparable](m map[T]bool) []T {
	out := make([]T, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
func singleLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(no description)"
	}
	a, _, found := strings.Cut(s, "\n")
	if found {
		return strings.TrimRight(a, " ") + " …"
	}
	return strings.TrimRight(a, " ")
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
