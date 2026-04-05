package main

import (
	"errors"
	"time"

	"github.com/zalando/go-keyring"
)

const KeyringService = "modelswitch"

var (
	ErrKeyringUnavailable = errors.New("system keyring unavailable")
	ErrSecretNotFound     = errors.New("secret not found in keyring")
)

type keyringResult struct {
	val string
	err error
}

func withTimeout(fn func() (string, error)) (string, error) {
	ch := make(chan keyringResult, 1)
	go func() {
		defer close(ch)
		val, err := fn()
		ch <- keyringResult{val, err}
	}()
	select {
	case res := <-ch:
		if errors.Is(res.err, keyring.ErrNotFound) {
			return "", ErrSecretNotFound
		}
		return res.val, res.err
	case <-time.After(3 * time.Second):
		return "", ErrKeyringUnavailable
	}
}

// KeyringSet stores a provider's API key in the system keyring.
func KeyringSet(provider, apiKey string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Set(KeyringService, provider, apiKey)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return ErrKeyringUnavailable
	}
}

// KeyringGet retrieves a provider's API key from the system keyring.
func KeyringGet(provider string) (string, error) {
	return withTimeout(func() (string, error) {
		return keyring.Get(KeyringService, provider)
	})
}

// KeyringDelete removes a provider's API key from the system keyring.
func KeyringDelete(provider string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Delete(KeyringService, provider)
	}()
	select {
	case err := <-ch:
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	case <-time.After(3 * time.Second):
		return ErrKeyringUnavailable
	}
}

// ResolveAPIKey returns the API key for a given provider.
// Checks keyring first (if UseKeyring), then falls back to config's APIKey.
func ResolveAPIKey(providerName string, p Provider) (string, error) {
	if p.UseKeyring {
		key, err := KeyringGet(providerName)
		if err != nil {
			return "", err
		}
		return key, nil
	}
	return p.APIKey, nil
}
