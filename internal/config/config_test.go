package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	_, err := Load("../../config")
	if err != nil {
		t.Fatal(err)
	}
}
