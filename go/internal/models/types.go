package models

import "time"

// Note represents a note file with metadata
type Note struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Folder   string   `json:"folder"`
	Modified string   `json:"modified"`
	Size     int64    `json:"size"`
	Type     string   `json:"type"`
	Tags     []string `json:"tags"`
}

// NoteMetadata represents metadata for a note
type NoteMetadata struct {
	Created  string `json:"created"`
	Modified string `json:"modified"`
	Size     int64  `json:"size"`
	Lines    int    `json:"lines"`
}

// NoteContent represents a note with its content
type NoteContent struct {
	Path     string       `json:"path"`
	Content  string       `json:"content"`
	Metadata NoteMetadata `json:"metadata"`
}

// MatchContext represents a search match with context
type MatchContext struct {
	LineNumber int    `json:"line_number"`
	Context    string `json:"context"`
}

// SearchResult represents a search result
type SearchResult struct {
	Name    string         `json:"name"`
	Path    string         `json:"path"`
	Folder  string         `json:"folder"`
	Type    string         `json:"type"`
	Matches []MatchContext `json:"matches"`
	Score   float64        `json:"score,omitempty"` // relevance score for smart mode
}

// SearchResultsResponse represents response from search requests
type SearchResultsResponse struct {
	Results    []SearchResult `json:"results"`
	Pagination PaginationMeta `json:"pagination,omitempty"`
}

// SearchResultsPaginated represents paginated search results
type SearchResultsPaginated struct {
	Results    []SearchResult `json:"results"`
	Pagination PaginationMeta `json:"pagination"`
}

// Template represents a note template
type Template struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Modified string `json:"modified"`
}

// Theme represents a UI theme
type Theme struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Builtin bool   `json:"builtin"`
}

// Locale represents a language locale
type Locale struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}

// ShareToken represents a share token for public note access
type ShareToken struct {
	Path      string `json:"path"`
	Theme     string `json:"theme"`
	Created   string `json:"created"`
	ExpiresAt string `json:"expires_at,omitempty"` // RFC3339, empty = never expires
}

// ShareInfo represents share information for a note
type ShareInfo struct {
	Shared  bool   `json:"shared"`
	Token   string `json:"token,omitempty"`
	Theme   string `json:"theme,omitempty"`
	Created string `json:"created,omitempty"`
	URL     string `json:"url,omitempty"`
}

// GraphNode represents a node in the knowledge graph
type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

	// GraphEdge represents an edge in the knowledge graph
type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

// GraphData represents the complete knowledge graph
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NoteSaveResponse represents response from saving a note
type NoteSaveResponse struct {
	Success  bool   `json:"success"`
	Path     string `json:"path"`
	Message  string `json:"message"`
	Content  string `json:"content"`
	Modified string `json:"modified"`
}

// NoteMoveResponse represents response from moving a note
type NoteMoveResponse struct {
	Success bool   `json:"success"`
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
	Message string `json:"message"`
}

// FolderResponse represents response from folder operations
type FolderResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

// MediaUploadResponse represents response from media upload
type MediaUploadResponse struct {
	Success  bool   `json:"success"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

// ShareCreateResponse represents response from creating a share
type ShareCreateResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	URL     string `json:"url"`
	Path    string `json:"path"`
	Theme   string `json:"theme"`
}

// ConfigResponse represents the public config for frontend
type ConfigResponse struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	SearchEnabled  bool   `json:"searchEnabled"`
	DemoMode       bool   `json:"demoMode"`
	AlreadyDonated bool   `json:"alreadyDonated"`
	Authentication struct {
		Enabled bool `json:"enabled"`
	} `json:"authentication"`
}

// NotesListResponse represents response from listing notes
type NotesListResponse struct {
	Notes      []Note         `json:"notes"`
	Folders    []string       `json:"folders"`
	Pagination PaginationMeta `json:"pagination,omitempty"`
}

// TagsResponse represents response from listing tags
type TagsResponse struct {
	Tags map[string]int `json:"tags"`
}

// TagNotesResponse represents response for notes by tag
type TagNotesResponse struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
	Notes []Note `json:"notes"`
}

// TemplatesResponse represents response from listing templates
type TemplatesResponse struct {
	Templates []Template `json:"templates"`
}

// ThemesResponse represents response from listing themes
type ThemesResponse struct {
	Themes []Theme `json:"themes"`
}

// ThemeResponse represents response for a single theme
type ThemeResponse struct {
	CSS     string `json:"css"`
	ThemeID string `json:"theme_id"`
}

// LocalesResponse represents response from listing locales
type LocalesResponse struct {
	Locales []Locale `json:"locales"`
}

// SharedNotesResponse represents response for shared notes list
type SharedNotesResponse struct {
	Paths []string `json:"paths"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status  string `json:"status"`
	App     string `json:"app"`
	Version string `json:"version"`
}

// ReadinessResponse represents detailed readiness check response
type ReadinessResponse struct {
	Status     string            `json:"status"`
	App        string            `json:"app"`
	Version    string            `json:"version"`
	Checks     map[string]string `json:"checks"`
}

// CacheEntry represents a cached scan result
type CacheEntry struct {
	CachedAt time.Time
	Notes    []Note
	Folders  []string
}

// TagCacheEntry represents a cached tag parse result
type TagCacheEntry struct {
	ModifiedTime time.Time
	Tags         []string
}


// OrphanedMediaFile represents an orphaned media file
type OrphanedMediaFile struct {
	Path      string `json:"path"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	MediaType string `json:"mediaType"`
	Type      string `json:"type"`
}

type Attachment struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

type AttachmentsResponse struct {
	Success     bool         `json:"success"`
	Attachments []Attachment `json:"attachments"`
	Count       int          `json:"count"`
}

// OrphanedMediaResponse represents response from listing orphaned media
type OrphanedMediaResponse struct {
	Success   bool                `json:"success"`
	Count     int                 `json:"count"`
	Files     []OrphanedMediaFile `json:"files"`
	TotalSize int64               `json:"totalSize"`
}

// CleanupMediaResponse represents response from cleanup operation
type CleanupMediaResponse struct {
	Success     bool     `json:"success"`
	DeletedCount int     `json:"deletedCount"`
	DeletedFiles []string `json:"deletedFiles"`
	FreedSpace  int64    `json:"freedSpace"`
	Message     string   `json:"message"`
}