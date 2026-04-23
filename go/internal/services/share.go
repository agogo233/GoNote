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

// loadTokens loads share tokens from file
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
		return make(map[string]models.ShareToken), nil
	}

	return tokens, nil
}

// saveTokens saves share tokens to file
func (s *ShareService) saveTokens(tokens map[string]models.ShareToken) error {
	tokensFile := s.getTokensFilePath()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(tokensFile), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokensFile, data, 0644)
}

// CreateShareToken creates a share token for a note
func (s *ShareService) CreateShareToken(notePath, theme string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.loadTokens()
	if err != nil {
		return "", err
	}

	// Check if note already has a token
	for token, info := range tokens {
		if info.Path == notePath {
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
	tokens[token] = models.ShareToken{
		Path:    notePath,
		Theme:   theme,
		Created: time.Now().UTC().Format(time.RFC3339),
	}

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
