package keyring

import (
	"errors"
	"fmt"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "dotkeeper"
	userName    = "backup-password"
)

var (
	ErrPasswordNotFound   = errors.New("password not found in keyring")
	ErrKeyringUnavailable = errors.New("keyring unavailable")
)

// Store stores the password in the system keyring
func Store(password string) error {
	err := keyring.Set(serviceName, userName, password)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrKeyringUnavailable, err)
	}
	return nil
}

// Retrieve retrieves the password from the system keyring
func Retrieve() (string, error) {
	password, err := keyring.Get(serviceName, userName)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", ErrPasswordNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrKeyringUnavailable, err)
	}
	return password, nil
}

// Delete removes the password from the system keyring
func Delete() error {
	err := keyring.Delete(serviceName, userName)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil // Already deleted is fine
		}
		return fmt.Errorf("%w: %v", ErrKeyringUnavailable, err)
	}
	return nil
}

// IsAvailable checks if the keyring is available
func IsAvailable() bool {
	// Try to store and retrieve a test value
	testValue := "test-" + fmt.Sprintf("%d", time.Now().UnixNano())
	err := keyring.Set(serviceName, userName+"-test", testValue)
	if err != nil {
		return false
	}

	// Clean up
	_ = keyring.Delete(serviceName, userName+"-test")
	return true
}
