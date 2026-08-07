package config

import "testing"

func TestDefaults(t *testing.T) {
	value := Defaults()
	if value.Stack.DefaultCommand != "log" {
		t.Fatalf("default stack command = %q", value.Stack.DefaultCommand)
	}
}
