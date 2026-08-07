package completion

import (
	"strings"
	"testing"
)

func TestWriteJJBridge(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		var output strings.Builder
		if err := WriteJJBridge(&output, shell); err != nil {
			t.Fatalf("WriteJJBridge(%q) error = %v", shell, err)
		}
		if !strings.Contains(output.String(), "bb") || !strings.Contains(output.String(), "stack") {
			t.Errorf("bridge for %s does not delegate to bb stack", shell)
		}
	}
}
