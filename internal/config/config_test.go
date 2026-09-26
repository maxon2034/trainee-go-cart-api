package config

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := Load("../../config")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(config.DB.Password)
}
