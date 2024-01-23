package main

import (
	"flag"
	"os"
	"strings"
)

// This file contains environnement variables parsing related methods,
// for configuration purpose.

// parseEnv parses configuration environnement variables.
func parseEnv() {
	if !isFlagSet("redis") {
		// Use either BYTEBOT_REDIS or REDIS_URL, but default to localhost:6379
		// BYTEBOT_REDIS is used by the Dockerfile and takes precedence if both are set
		if os.Getenv("REDIS_URL") != "" {
			*redisAddr = os.Getenv("REDIS_URL")
		} else {
			*redisAddr = "localhost:6379"
		}
	}

	if !isFlagSet("id") {
		*id = parseStringFromEnv("BYTEBOT_ID", "discord")
	}

	if !isFlagSet("t") {
		Token = parseStringFromEnv("BYTEBOT_TOKEN", "")
	}

	if !isFlagSet("rpass") {
		*redisPass = parseStringFromEnv("BYTEBOT_RPASS", "")
	}

	if !isFlagSet("ruser") {
		*redisUser = parseStringFromEnv("BYTEBOT_RUSER", "")
	}

	if !isFlagSet("verbose") {
		*verbose = parseBoolFromEnv("BYTEBOT_VERBOSE", false)
	}
}

// Parses a string from an env variable and returns it.
func parseStringFromEnv(varName, defaultVal string) string {
	val, set := os.LookupEnv(varName)
	if set {
		return val
	}
	return defaultVal
}

// Parses a boolean from an env variable and returns it.
func parseBoolFromEnv(varName string, defaultVal bool) bool {
	val, set := os.LookupEnv(varName)
	if set {
		return strings.ToLower(val) == "true"
	}
	return false
}

// This is used to check if a flag was set
// Must be called after flag.Parse()
func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
