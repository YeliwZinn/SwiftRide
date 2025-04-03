package config

import (
    "os"
    "log"
)

func GetEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func MustGetEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        log.Fatalf("Environment variable %s not set", key)
    }
    return value
}