package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func GetOptionalString(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Printf("Could not find the value for the key '%s'. Using default value '%s'", key, def)
		return def
	}
	return val
}

func GetMandatoryString(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Could not find the value for the key '%s'", key)
	}
	return val
}

func GetBoolean(key string) bool {
	text := os.Getenv(key)
	if text == "" {
		return false
	}

	val, err := strconv.ParseBool(text)
	if err != nil {
		log.Fatalf("Could not parse value '%s' as boolean", text)
	}

	return val
}
