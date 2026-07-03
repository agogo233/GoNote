package services

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestShareService_CreateShareToken(t *testing.T) {
	t.Run("create new token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, err := svc.CreateShareToken("notes/test.md", "dark")
		if err != nil {
			t.Fatalf("CreateShareToken failed: %v", err)
		}
		if len(token) != 32 {
			t.Errorf("Expected token length 32, got %d", len(token))
		}
	})

	t.Run("same note returns same token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token1, err := svc.CreateShareToken("notes/test.md", "dark")
		if err != nil {
			t.Fatalf("First CreateShareToken failed: %v", err)
		}

		token2, err := svc.CreateShareToken("notes/test.md", "light")
		if err != nil {
			t.Fatalf("Second CreateShareToken failed: %v", err)
		}

		if token1 != token2 {
			t.Errorf("Expected same token for same note, got %s and %s", token1, token2)
		}
	})

	t.Run("different notes get different tokens", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token1, _ := svc.CreateShareToken("notes/note1.md", "dark")
		token2, _ := svc.CreateShareToken("notes/note2.md", "dark")

		if token1 == token2 {
			t.Error("Different notes should have different tokens")
		}
	})

	t.Run("token persisted to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token1, _ := svc.CreateShareToken("notes/test.md", "dark")

		// Create new service instance (should load from file)
		svc2 := NewShareService(tmpDir)
		token2, exists := svc2.GetShareToken("notes/test.md")

		if !exists {
			t.Error("Token should persist across service instances")
		}
		if token1 != token2 {
			t.Error("Token should be the same after reload")
		}
	})

	t.Run("create token with chinese path", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, err := svc.CreateShareToken("笔记/测试笔记.md", "dark")
		if err != nil {
			t.Fatalf("CreateShareToken with chinese path failed: %v", err)
		}
		if len(token) != 32 {
			t.Errorf("Expected token length 32, got %d", len(token))
		}
	})
}

func TestShareService_GetNoteByToken(t *testing.T) {
	t.Run("non-existent token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		info, exists := svc.GetNoteByToken("nonexistent")
		if exists {
			t.Error("Expected not exists for non-existent token")
		}
		if info != nil {
			t.Error("Expected nil info for non-existent token")
		}
	})

	t.Run("existing token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, _ := svc.CreateShareToken("notes/test.md", "dark")
		info, exists := svc.GetNoteByToken(token)

		if !exists {
			t.Fatal("Expected token to exist")
		}
		if info.Path != "notes/test.md" {
			t.Errorf("Expected path 'notes/test.md', got %q", info.Path)
		}
		if info.Theme != "dark" {
			t.Errorf("Expected theme 'dark', got %q", info.Theme)
		}
		if info.Created == "" {
			t.Error("Expected Created to be set")
		}
	})
}

func TestShareService_GetShareToken(t *testing.T) {
	t.Run("note without token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, exists := svc.GetShareToken("notes/unshared.md")
		if exists {
			t.Error("Expected not exists for unshared note")
		}
		if token != "" {
			t.Error("Expected empty token for unshared note")
		}
	})

	t.Run("note with token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		createdToken, _ := svc.CreateShareToken("notes/shared.md", "light")
		token, exists := svc.GetShareToken("notes/shared.md")

		if !exists {
			t.Error("Expected exists for shared note")
		}
		if token != createdToken {
			t.Errorf("Expected token %s, got %s", createdToken, token)
		}
	})
}

func TestShareService_RevokeShareToken(t *testing.T) {
	t.Run("revoke existing token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, _ := svc.CreateShareToken("notes/test.md", "dark")
		err := svc.RevokeShareToken("notes/test.md")

		if err != nil {
			t.Fatalf("RevokeShareToken failed: %v", err)
		}

		_, exists := svc.GetNoteByToken(token)
		if exists {
			t.Error("Token should not exist after revocation")
		}
	})

	t.Run("revoke non-existent token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		err := svc.RevokeShareToken("notes/nonexistent.md")
		if err == nil {
			t.Error("Expected error for non-existent token")
		}
	})

	t.Run("revoke persists across instances", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, _ := svc.CreateShareToken("notes/test.md", "dark")
		svc.RevokeShareToken("notes/test.md")

		// New instance
		svc2 := NewShareService(tmpDir)
		_, exists := svc2.GetNoteByToken(token)
		if exists {
			t.Error("Token should be revoked in new instance")
		}
	})
}

func TestShareService_GetAllSharedPaths(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		paths, err := svc.GetAllSharedPaths()
		if err != nil {
			t.Fatalf("GetAllSharedPaths failed: %v", err)
		}
		if len(paths) != 0 {
			t.Errorf("Expected empty list, got %d paths", len(paths))
		}
	})

	t.Run("multiple shared paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		svc.CreateShareToken("notes/note1.md", "dark")
		svc.CreateShareToken("notes/note2.md", "light")
		svc.CreateShareToken("folder/note3.md", "nord")

		paths, err := svc.GetAllSharedPaths()
		if err != nil {
			t.Fatalf("GetAllSharedPaths failed: %v", err)
		}
		if len(paths) != 3 {
			t.Errorf("Expected 3 paths, got %d", len(paths))
		}
	})
}

func TestShareService_UpdateTokenPath(t *testing.T) {
	t.Run("update existing token path", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, _ := svc.CreateShareToken("old/path.md", "dark")

		err := svc.UpdateTokenPath("old/path.md", "new/path.md")
		if err != nil {
			t.Fatalf("UpdateTokenPath failed: %v", err)
		}

		info, _ := svc.GetNoteByToken(token)
		if info.Path != "new/path.md" {
			t.Errorf("Expected path 'new/path.md', got %q", info.Path)
		}
	})

	t.Run("update non-existent path", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		// Should not error, just do nothing
		err := svc.UpdateTokenPath("nonexistent.md", "new/path.md")
		if err != nil {
			t.Fatalf("UpdateTokenPath for non-existent should not error: %v", err)
		}
	})

	t.Run("update persists across instances", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, _ := svc.CreateShareToken("old/path.md", "dark")
		svc.UpdateTokenPath("old/path.md", "new/path.md")

		// New instance
		svc2 := NewShareService(tmpDir)
		info, _ := svc2.GetNoteByToken(token)
		if info.Path != "new/path.md" {
			t.Errorf("Updated path should persist, got %q", info.Path)
		}
	})
}

func TestShareService_DeleteTokenForNote(t *testing.T) {
	t.Run("delete existing token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, _ := svc.CreateShareToken("notes/test.md", "dark")
		err := svc.DeleteTokenForNote("notes/test.md")

		if err != nil {
			t.Fatalf("DeleteTokenForNote failed: %v", err)
		}

		_, exists := svc.GetNoteByToken(token)
		if exists {
			t.Error("Token should be deleted")
		}
	})

	t.Run("delete non-existent token", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		err := svc.DeleteTokenForNote("nonexistent.md")
		if err == nil {
			t.Error("Expected error for non-existent token")
		}
	})
}

func TestShareService_GetShareInfo(t *testing.T) {
	t.Run("shared note info", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, _ := svc.CreateShareToken("notes/test.md", "gruvbox-dark")
		info, err := svc.GetShareInfo("notes/test.md")

		if err != nil {
			t.Fatalf("GetShareInfo failed: %v", err)
		}
		if !info.Shared {
			t.Error("Expected Shared to be true")
		}
		if info.Token != token {
			t.Errorf("Expected token %s, got %s", token, info.Token)
		}
		if info.Theme != "gruvbox-dark" {
			t.Errorf("Expected theme 'gruvbox-dark', got %s", info.Theme)
		}
	})

	t.Run("unshared note info", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		info, err := svc.GetShareInfo("notes/unshared.md")

		if err != nil {
			t.Fatalf("GetShareInfo failed: %v", err)
		}
		if info.Shared {
			t.Error("Expected Shared to be false")
		}
		if info.Token != "" {
			t.Error("Expected empty token for unshared note")
		}
	})
}

func TestGenerateToken(t *testing.T) {
	t.Run("token length", func(t *testing.T) {
		token, err := generateToken(16)
		if err != nil {
			t.Fatalf("generateToken failed: %v", err)
		}
		if len(token) != 16 {
			t.Errorf("Expected length 16, got %d", len(token))
		}
	})

	t.Run("token uniqueness", func(t *testing.T) {
		tokens := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			token, err := generateToken(16)
			if err != nil {
				t.Fatalf("generateToken failed: %v", err)
			}
			if tokens[token] {
				t.Errorf("Duplicate token generated: %s", token)
			}
			tokens[token] = true
		}
	})

	t.Run("different lengths", func(t *testing.T) {
		lengths := []int{8, 16, 32, 64}
		for _, length := range lengths {
			token, err := generateToken(length)
			if err != nil {
				t.Fatalf("generateToken(%d) failed: %v", length, err)
			}
			if len(token) != length {
				t.Errorf("Expected length %d, got %d", length, len(token))
			}
		}
	})

	t.Run("token characters", func(t *testing.T) {
		token, _ := generateToken(16)
		// Token should be hex characters only
		for _, c := range token {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("Invalid character in token: %c", c)
			}
		}
	})
}

func TestShareService_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent token creation", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		var wg sync.WaitGroup
		numGoroutines := 50
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := svc.CreateShareToken(fmt.Sprintf("note%d.md", id), "dark")
				if err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Errorf("Concurrent CreateShareToken failed: %v", err)
		}

		// Verify all tokens were stored
		paths, _ := svc.GetAllSharedPaths()
		if len(paths) != numGoroutines {
			t.Errorf("Expected %d paths, got %d", numGoroutines, len(paths))
		}
	})

	t.Run("concurrent read and write", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		// Pre-create some tokens
		for i := 0; i < 10; i++ {
			svc.CreateShareToken(fmt.Sprintf("note%d.md", i), "dark")
		}

		var wg sync.WaitGroup

		// Concurrent reads
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				svc.GetShareToken(fmt.Sprintf("note%d.md", id%10))
			}(i)
		}

		// Concurrent writes
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				svc.CreateShareToken(fmt.Sprintf("newnote%d.md", id), "light")
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent create and revoke", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		var wg sync.WaitGroup

		// Create tokens
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				svc.CreateShareToken(fmt.Sprintf("note%d.md", id), "dark")
			}(i)
		}

		wg.Wait()

		// Concurrent revoke
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				svc.RevokeShareToken(fmt.Sprintf("note%d.md", id))
			}(i)
		}

		wg.Wait()

		// Should have 10 remaining
		paths, _ := svc.GetAllSharedPaths()
		if len(paths) != 10 {
			t.Errorf("Expected 10 remaining paths, got %d", len(paths))
		}
	})
}

func TestShareService_CorruptedTokenFile(t *testing.T) {
	t.Run("corrupted json file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Write invalid JSON
		tokensFile := fmt.Sprintf("%s/.share-tokens.json", tmpDir)
		os.WriteFile(tokensFile, []byte("invalid json {{{"), 0644)

		// Should not crash, just return empty
		svc := NewShareService(tmpDir)
		paths, err := svc.GetAllSharedPaths()
		if err != nil {
			t.Fatalf("GetAllSharedPaths should not error: %v", err)
		}
		if len(paths) != 0 {
			t.Errorf("Expected empty paths for corrupted file, got %d", len(paths))
		}
	})

	t.Run("missing tokens file", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Don't create the file

		svc := NewShareService(tmpDir)
		paths, err := svc.GetAllSharedPaths()
		if err != nil {
			t.Fatalf("GetAllSharedPaths should not error: %v", err)
		}
		if len(paths) != 0 {
			t.Errorf("Expected empty paths for missing file, got %d", len(paths))
		}
	})
}

func TestShareService_CreateShareTokenWithTTL(t *testing.T) {
	t.Run("expired token is filtered on load", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		// Create a permanent token first, then rewrite its ExpiresAt to the past
		// and reload to verify loadTokens filters expired entries.
		token, err := svc.CreateShareToken("notes/test.md", "dark")
		if err != nil {
			t.Fatalf("CreateShareToken failed: %v", err)
		}

		// Rewrite the tokens file with an already-expired ExpiresAt.
		tokensFile := svc.getTokensFilePath()
		past := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
		corpus := fmt.Sprintf(`{"%s":{"path":"notes/test.md","theme":"dark","created":"2024-01-01T00:00:00Z","expires_at":"%s"}}`, token, past)
		if err := os.WriteFile(tokensFile, []byte(corpus), 0644); err != nil {
			t.Fatalf("rewrite failed: %v", err)
		}

		// Fresh service loads from disk and should filter out the expired token.
		svc2 := NewShareService(tmpDir)
		_, exists := svc2.GetNoteByToken(token)
		if exists {
			t.Error("Expired token should not be accessible via GetNoteByToken")
		}
		_, exists = svc2.GetShareToken("notes/test.md")
		if exists {
			t.Error("Expired token should not be returned via GetShareToken")
		}
	})

	t.Run("future-expiry token is accessible", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, err := svc.CreateShareTokenWithTTL("notes/live.md", "light", 1*time.Hour)
		if err != nil {
			t.Fatalf("CreateShareTokenWithTTL failed: %v", err)
		}

		info, exists := svc.GetNoteByToken(token)
		if !exists {
			t.Fatal("Future-expiry token should be accessible")
		}
		if info.ExpiresAt == "" {
			t.Error("Expected ExpiresAt to be set for TTL token")
		}
	})

	t.Run("ttl zero means never expires", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		token, err := svc.CreateShareTokenWithTTL("notes/forever.md", "light", 0)
		if err != nil {
			t.Fatalf("CreateShareTokenWithTTL failed: %v", err)
		}
		info, exists := svc.GetNoteByToken(token)
		if !exists {
			t.Fatal("Zero-TTL token should be accessible")
		}
		if info.ExpiresAt != "" {
			t.Errorf("Expected empty ExpiresAt for never-expire token, got %q", info.ExpiresAt)
		}
	})

	t.Run("existing token refresh expiry when ttl given", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := NewShareService(tmpDir)

		// Create permanent token.
		token, _ := svc.CreateShareToken("notes/refresh.md", "dark")
		// Re-create with TTL: should NOT create new token, but refresh expiry.
		token2, err := svc.CreateShareTokenWithTTL("notes/refresh.md", "dark", 2*time.Hour)
		if err != nil {
			t.Fatalf("refresh failed: %v", err)
		}
		if token != token2 {
			t.Fatalf("Expected same token after refresh, got %s vs %s", token, token2)
		}
		info, _ := svc.GetNoteByToken(token)
		if info.ExpiresAt == "" {
			t.Error("Expected ExpiresAt to be refreshed")
		}
	})
}

func TestShareService_CorruptedTokenFileBackedUp(t *testing.T) {
	tmpDir := t.TempDir()

	tokensFile := fmt.Sprintf("%s/.share-tokens.json", tmpDir)
	os.WriteFile(tokensFile, []byte("invalid json {{{"), 0644)

	svc := NewShareService(tmpDir)
	// Service should still be usable.
	_, err := svc.GetAllSharedPaths()
	if err != nil {
		t.Fatalf("GetAllSharedPaths should not error on corrupt file: %v", err)
	}

	// The corrupt file should have been renamed to a .broken.<unix> backup.
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".share-tokens.json.broken.") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected a backup file (.share-tokens.json.broken.*) to exist after corrupt load")
	}

	// And we can now create a fresh token (write-path works).
	_, err = svc.CreateShareToken("notes/after.md", "dark")
	if err != nil {
		t.Fatalf("CreateShareToken after corrupt-recovery failed: %v", err)
	}
}

func TestAtomicWrite(t *testing.T) {
	t.Run("writes content atomically", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := tmpDir + "/sub/file.txt"
		want := []byte("hello world")
		if err := AtomicWrite(path, want, 0644); err != nil {
			t.Fatalf("AtomicWrite failed: %v", err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(got) != string(want) {
			t.Errorf("got %q want %q", got, want)
		}
		// No leftover .tmp files in the directory.
		entries, _ := os.ReadDir(tmpDir + "/sub")
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".tmp") {
				t.Errorf("leftover temp file: %s", e.Name())
			}
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := tmpDir + "/f.txt"
		os.WriteFile(path, []byte("old"), 0644)
		if err := AtomicWrite(path, []byte("new content"), 0644); err != nil {
			t.Fatalf("AtomicWrite failed: %v", err)
		}
		got, _ := os.ReadFile(path)
		if string(got) != "new content" {
			t.Errorf("got %q want %q", got, "new content")
		}
	})

	t.Run("preserves file mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := tmpDir + "/mode.txt"
		if err := AtomicWrite(path, []byte("x"), 0600); err != nil {
			t.Fatalf("AtomicWrite failed: %v", err)
		}
		info, _ := os.Stat(path)
		if info.Mode().Perm() != 0600 {
			t.Errorf("got mode %v want 0600", info.Mode().Perm())
		}
	})
}
