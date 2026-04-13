package storage

import (
	"shminer/backend/types"
	"testing"
)

func setupTestEnv(t *testing.T) {
	tempDir := t.TempDir()

	t.Setenv("APPDATA", tempDir)         // Windows
	t.Setenv("XDG_CONFIG_HOME", tempDir) // Linux
	t.Setenv("HOME", tempDir)            // macOS та fallback для Linux
}

func TestStorage_InitAndLoad(t *testing.T) {
	setupTestEnv(t)

	s := NewDriver()
	err := s.InitStorage("my_secure_password")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	if !s.CheckExists() {
		t.Fatalf("Storage file should exist")
	}

	s2 := NewDriver()
	err = s2.LoadStorage("my_secure_password")
	if err != nil {
		t.Fatalf("Failed to load storage: %v", err)
	}

	if s2.GetSessionPassword() != "my_secure_password" {
		t.Errorf("Expected session password 'my_secure_password', got '%s'", s2.GetSessionPassword())
	}
}

func TestStorage_ChangePassword(t *testing.T) {
	setupTestEnv(t)

	s := NewDriver()
	s.InitStorage("old_password")

	err := s.ChangePassword("wrong_password", "new_password")
	if err == nil {
		t.Errorf("Expected error when changing password with wrong old password")
	}

	err = s.ChangePassword("old_password", "new_password")
	if err != nil {
		t.Fatalf("Failed to change password: %v", err)
	}

	if s.GetSessionPassword() != "new_password" {
		t.Errorf("Session password should be updated to 'new_password'")
	}

	// Verify we can load with new password
	s2 := NewDriver()
	err = s2.LoadStorage("new_password")
	if err != nil {
		t.Fatalf("Failed to load with new password: %v", err)
	}
}

func TestStorage_UpdateWallets(t *testing.T) {
	setupTestEnv(t)

	s := NewDriver()
	s.InitStorage("test")

	wallets := []types.WalletStats{
		{Address: "addr1", Name: "Wallet 1"},
	}

	s.UpdateWallets(wallets)

	data := s.GetStorage()
	if len(data.Wallets) != 1 || data.Wallets[0].Address != "addr1" {
		t.Errorf("Failed to update wallets")
	}
}
