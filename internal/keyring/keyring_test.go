package keyring

import (
	"testing"
)

func TestKeyringRoundtrip(t *testing.T) {
	// Skip if keyring is not available
	if !IsAvailable() {
		t.Skip("keyring not available on this system")
	}

	// Clean up before test
	_ = Delete()

	testPassword := "test-password-12345"

	// Test Store
	err := Store(testPassword)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Test Retrieve
	retrieved, err := Retrieve()
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if retrieved != testPassword {
		t.Errorf("Retrieved password mismatch: got %q, want %q", retrieved, testPassword)
	}

	// Test Delete
	err = Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = Retrieve()
	if err != ErrPasswordNotFound {
		t.Errorf("Expected ErrPasswordNotFound after deletion, got %v", err)
	}

	// Clean up after test
	_ = Delete()
}

func TestRetrieveNotFound(t *testing.T) {
	// Skip if keyring is not available
	if !IsAvailable() {
		t.Skip("keyring not available on this system")
	}

	// Clean up before test
	_ = Delete()

	// Try to retrieve non-existent password
	_, err := Retrieve()
	if err != ErrPasswordNotFound {
		t.Errorf("Expected ErrPasswordNotFound, got %v", err)
	}

	// Clean up after test
	_ = Delete()
}

func TestDeleteNonExistent(t *testing.T) {
	// Skip if keyring is not available
	if !IsAvailable() {
		t.Skip("keyring not available on this system")
	}

	// Clean up before test
	_ = Delete()

	// Delete non-existent password should not error
	err := Delete()
	if err != nil {
		t.Errorf("Delete non-existent should not error, got %v", err)
	}
}
