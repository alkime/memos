// Package keyring provides access to the system keychain for storing API keys.
package keyring

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const serviceName = "memos-voice"

// Key represents a named API key stored in the keychain.
type Key string

const (
	// OpenAIKey is the keychain entry for the OpenAI API key.
	OpenAIKey Key = "openai-api-key"
	// AnthropicKey is the keychain entry for the Anthropic API key.
	AnthropicKey Key = "anthropic-api-key"
)

// AllKeys returns all known key types for iteration.
func AllKeys() []Key {
	return []Key{OpenAIKey, AnthropicKey}
}

// DisplayName returns a human-readable name for the key.
func (k Key) DisplayName() string {
	switch k {
	case OpenAIKey:
		return "openai"
	case AnthropicKey:
		return "anthropic"
	default:
		return string(k)
	}
}

// Get retrieves a key value from the system keychain.
func Get(key Key) (string, error) {
	value, err := keyring.Get(serviceName, string(key))
	if err != nil {
		return "", fmt.Errorf("failed to get %s from keychain: %w", key.DisplayName(), err)
	}

	return value, nil
}

// Set stores a key value in the system keychain.
func Set(key Key, value string) error {
	if err := keyring.Set(serviceName, string(key), value); err != nil {
		return fmt.Errorf("failed to set %s in keychain: %w", key.DisplayName(), err)
	}

	return nil
}

// IsSet checks if a key exists in the keychain.
func IsSet(key Key) bool {
	_, err := keyring.Get(serviceName, string(key))

	return err == nil
}

// KeyFromServiceName maps a service name (e.g., "openai") to a Key.
func KeyFromServiceName(name string) (Key, error) {
	switch name {
	case "openai":
		return OpenAIKey, nil
	case "anthropic":
		return AnthropicKey, nil
	default:
		return "", fmt.Errorf("unknown service: %s", name)
	}
}
