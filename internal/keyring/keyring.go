// Package keyring provides access to the system keychain for storing API keys.
package keyring

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const serviceName = "memos-voice"

// APIKey represents a named API key stored in the keychain.
type APIKey string

const (
	// OpenAI is the keychain entry for the OpenAI API key.
	OpenAI APIKey = "openai-api-key"
	// Anthropic is the keychain entry for the Anthropic API key.
	Anthropic APIKey = "anthropic-api-key"
)

// AllAPIKeys returns all known API key types for iteration.
func AllAPIKeys() []APIKey {
	return []APIKey{OpenAI, Anthropic}
}

// DisplayName returns a human-readable name for the API key.
func (k APIKey) DisplayName() string {
	switch k {
	case OpenAI:
		return "openai"
	case Anthropic:
		return "anthropic"
	default:
		return string(k)
	}
}

// Get retrieves an API key value from the system keychain.
func Get(apiKey APIKey) (string, error) {
	value, err := keyring.Get(serviceName, string(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to get %s from keychain: %w", apiKey.DisplayName(), err)
	}

	return value, nil
}

// Set stores an API key value in the system keychain.
func Set(apiKey APIKey, value string) error {
	if err := keyring.Set(serviceName, string(apiKey), value); err != nil {
		return fmt.Errorf("failed to set %s in keychain: %w", apiKey.DisplayName(), err)
	}

	return nil
}

// IsSet checks if an API key exists in the keychain.
func IsSet(apiKey APIKey) bool {
	_, err := keyring.Get(serviceName, string(apiKey))

	return err == nil
}

// APIKeyFromServiceName maps a service name (e.g., "openai") to an APIKey.
func APIKeyFromServiceName(name string) (APIKey, error) {
	switch name {
	case "openai":
		return OpenAI, nil
	case "anthropic":
		return Anthropic, nil
	default:
		return "", fmt.Errorf("unknown service: %s", name)
	}
}
