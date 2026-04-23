package services

import (
	"math"

	"gonote/internal/models"
)

const (
	DefaultLimit = 50
	MaxLimit     = 10000 // Increased to support tag filtering across all notes
)

func Paginate(notes []models.Note, page, limit int) models.PaginatedResult {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = DefaultLimit
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}

	total := len(notes)
	if total == 0 {
		return models.PaginatedResult{
			Notes: []models.Note{},
			Pagination: models.PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      0,
				TotalPages: 0,
				HasNext:    false,
				HasPrev:    false,
			},
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	if page > totalPages && totalPages > 0 {
		return models.PaginatedResult{
			Notes: []models.Note{},
			Pagination: models.PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      total,
				TotalPages: totalPages,
				HasNext:    false,
				HasPrev:    page > 1,
			},
		}
	}

	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}

	paginatedNotes := notes[start:end]

	return models.PaginatedResult{
		Notes: paginatedNotes,
		Pagination: models.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}
}

// PaginateSearchResults paginates search results
func PaginateSearchResults(results []models.SearchResult, page, limit int) models.SearchResultsPaginated {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = DefaultLimit
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}

	total := len(results)
	if total == 0 {
		return models.SearchResultsPaginated{
			Results: []models.SearchResult{},
			Pagination: models.PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      0,
				TotalPages: 0,
				HasNext:    false,
				HasPrev:    false,
			},
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	if page > totalPages && totalPages > 0 {
		return models.SearchResultsPaginated{
			Results: []models.SearchResult{},
			Pagination: models.PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      total,
				TotalPages: totalPages,
				HasNext:    false,
				HasPrev:    page > 1,
			},
		}
	}

	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}

	paginatedResults := results[start:end]

	return models.SearchResultsPaginated{
		Results: paginatedResults,
		Pagination: models.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}
}
