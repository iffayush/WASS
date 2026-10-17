package engine

import "testing"

func TestVersionIsSet(t *testing.T) {
	if Version == "" {
		t.Fatal("engine.Version must not be empty")
	}
}
