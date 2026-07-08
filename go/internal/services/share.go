package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gonote/internal/models"
	"gonote/internal/models/logger"
)

// ShareService handles share token operations
type ShareService struct {
	notesDir string
	mu       sync.RWMutex
}

// NewShareService creates a new ShareService
func NewShareService(notesDir string) *ShareService {
	return &ShareService{notesDir: notesDir}
}

// generateToken generates a URL-safe random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// getTokensFilePath returns the path to the share tokens file
func (s *ShareService) getTokensFilePath() string {
	return filepath.Join(s.notesDir, ".share-tokens.json")
}

// loadTokens loads share tokens from file. Expired tokens are filtered out
// (but not persisted here — they will be dropped on the next successful
// mutation that calls saveTokens). A corrupted token file is backed up to
// .share-tokens.json.broken.<unix> and an empty map is returned with a nil
// error so the service stays usable, instead of silently wiping the tokens
// (or fataling the startup).
func (s *ShareService) loadTokens() (map[string]models.ShareToken, error) {
	tokensFile := s.getTokensFilePath()

	data, err := os.ReadFile(tokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]models.ShareToken), nil
		}
		return nil, err
	}

	var tokens map[string]models.ShareToken
	if err := json.Unmarshal(data, &tokens); err != nil {
		// Backup the corrupt file with a timestamped name so the operator can
		// investigate, then fall back to an empty token set rather than
		// silently losing all share links.
		backup := fmt.Sprintf("%s.broken.%d", tokensFile, time.Now().Unix())
		if renameErr := os.Rename(tokensFile, backup); renameErr != nil {
			logger.Errorf("[share] tokens file corrupt AND backup failed: %v (orig error: %v)", renameErr, err)
		} else {
			logger.Errorf("[share] tokens file corrupt, backed up to %s (error: %v)", backup, err)
		}
		return make(map[string]models.ShareToken), nil
	}

	// Filter out expired tokens in-memory. Persistence cleanup happens lazily
	// on the next saveTokens call.
	now := time.Now().UTC()
	for token, info := range tokens {
		if info.ExpiresAt == "" {
			continue
		}
		expiresAt, parseErr := time.Parse(time.RFC3339, info.ExpiresAt)
		if parseErr != nil {
			// Unparseable expiry: treat as expired to be safe.
			delete(tokens, token)
			continue
		}
		if now.After(expiresAt) {
			delete(tokens, token)
		}
	}

	return tokens, nil
}

// saveTokens saves share tokens to file atomically so a crash mid-write
// cannot leave a half-written token file.
func (s *ShareService) saveTokens(tokens map[string]models.ShareToken) error {
	tokensFile := s.getTokensFilePath()

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	return AtomicWrite(tokensFile, data, 0644)
}

// CreateShareToken creates a share token for a note with no expiry.
// Backward-compatible wrapper around CreateShareTokenWithTTL.
func (s *ShareService) CreateShareToken(notePath, theme string) (string, error) {
	return s.CreateShareTokenWithTTL(notePath, theme, 0)
}

// CreateShareTokenWithTTL creates a share token for a note with an optional
// time-to-live. A ttl <= 0 means the token never expires.
func (s *ShareService) CreateShareTokenWithTTL(notePath, theme string, ttl time.Duration) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.loadTokens()
	if err != nil {
		return "", err
	}

	// Check if note already has a token (and refresh its expiry if ttl given)
	for token, info := range tokens {
		if info.Path == notePath {
			if ttl > 0 {
				info.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
				tokens[token] = info
				if saveErr := s.saveTokens(tokens); saveErr != nil {
					return "", saveErr
				}
			}
			return token, nil
		}
	}

	// Generate new token (32 characters = 16 bytes for better security)
	token, err := generateToken(32)
	if err != nil {
		return "", err
	}

	// Ensure uniqueness
	for _, exists := tokens[token]; exists; {
		token, err = generateToken(32)
		if err != nil {
			return "", err
		}
		_, exists = tokens[token]
	}

	// Store token
	st := models.ShareToken{
		Path:    notePath,
		Theme:   theme,
		Created: time.Now().UTC().Format(time.RFC3339),
	}
	if ttl > 0 {
		st.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
	}
	tokens[token] = st

	if err := s.saveTokens(tokens); err != nil {
		return "", err
	}

	return token, nil
}

// GetShareToken gets the share token for a note
func (s *ShareService) GetShareToken(notePath string) (string, bool) {
	tokens, err := s.loadTokens()
	if err != nil {
		return "", false
	}

	for token, info := range tokens {
		if info.Path == notePath {
			return token, true
		}
	}

	return "", false
}

// GetNoteByToken returns note info for a share token
func (s *ShareService) GetNoteByToken(token string) (*models.ShareToken, bool) {
	tokens, err := s.loadTokens()
	if err != nil {
		return nil, false
	}

	info, exists := tokens[token]
	if !exists {
		return nil, false
	}

	return &info, true
}

// GetAllSharedPaths returns all shared note paths
func (s *ShareService) GetAllSharedPaths() ([]string, error) {
	tokens, err := s.loadTokens()
	if err != nil {
		return nil, err
	}

	paths := []string{}
	for _, info := range tokens {
		if info.Path != "" {
			paths = append(paths, info.Path)
		}
	}

	return paths, nil
}

// RevokeShareToken revokes the share token for a note
func (s *ShareService) RevokeShareToken(notePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.loadTokens()
	if err != nil {
		return err
	}

	// Find and remove token for this note
	tokenToRemove := ""
	for token, info := range tokens {
		if info.Path == notePath {
			tokenToRemove = token
			break
		}
	}

	if tokenToRemove != "" {
		delete(tokens, tokenToRemove)
		return s.saveTokens(tokens)
	}

	return fmt.Errorf("no share token found for note")
}

// GetShareInfo returns share information for a note
func (s *ShareService) GetShareInfo(notePath string) (*models.ShareInfo, error) {
	tokens, err := s.loadTokens()
	if err != nil {
		return nil, err
	}

	for token, info := range tokens {
		if info.Path == notePath {
			return &models.ShareInfo{
				Shared:  true,
				Token:   token,
				Theme:   info.Theme,
				Created: info.Created,
			}, nil
		}
	}

	return &models.ShareInfo{Shared: false}, nil
}

// UpdateTokenPath updates the path for a token when a note is moved
func (s *ShareService) UpdateTokenPath(oldPath, newPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.loadTokens()
	if err != nil {
		return err
	}

	for token, info := range tokens {
		if info.Path == oldPath {
			info.Path = newPath
			tokens[token] = info
			return s.saveTokens(tokens)
		}
	}

	return nil
}

// DeleteTokenForNote deletes the share token when a note is deleted
func (s *ShareService) DeleteTokenForNote(notePath string) error {
	return s.RevokeShareToken(notePath)
}
