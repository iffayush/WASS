package server

import "testing"

func TestVersionIsSet(t *testing.T) {
	if Version == "" {
		t.Fatal("server.Version must not be empty")
	}
}
