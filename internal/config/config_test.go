package config

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
)

func TestLoadConfig(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal("Error loading .env file")
	}
	config, err := LoadConfig("../../config")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(config)
}
