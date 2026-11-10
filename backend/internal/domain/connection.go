package domain

import (
	"net/url"
	"regexp"
	"strings"
)

const (
	MaxConnectionNameLength = 100
	MaxProfileNameLength    = 128
)

var regionPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)

// Connection is a saved, non-secret DynamoDB connection configuration.
type Connection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Profile  string `json:"profile,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type ConnectionInput struct {
	Name     string
	Region   string
	Profile  string
	Endpoint string
}

type ConnectionSettings struct {
	Region   string
	Profile  string
	Endpoint string
}

func ValidateConnectionInput(input ConnectionInput) (ConnectionInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Region = strings.TrimSpace(input.Region)
	input.Profile = strings.TrimSpace(input.Profile)
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	if input.Name == "" || len(input.Name) > MaxConnectionNameLength || hasControl(input.Name) {
		return ConnectionInput{}, InvalidInput("Name must be between 1 and 100 characters")
	}
	if !regionPattern.MatchString(input.Region) || !strings.Contains(input.Region, "-") {
		return ConnectionInput{}, InvalidInput("Region is invalid")
	}
	if len(input.Profile) > MaxProfileNameLength || hasControl(input.Profile) {
		return ConnectionInput{}, InvalidInput("Profile must be at most 128 characters")
	}
	if input.Endpoint != "" {
		parsed, err := url.ParseRequestURI(input.Endpoint)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
			return ConnectionInput{}, InvalidInput("Endpoint must be an absolute HTTP(S) URL without credentials, path, query, or fragment")
		}
	}
	return input, nil
}

func hasControl(value string) bool {
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}
