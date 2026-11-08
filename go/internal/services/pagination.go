package services

import (
	"math"

	"gonote/internal/models"
)

const (
	DefaultLimit = 50
	MaxLimit     = 10000 // Increased to support tag filtering across all notes
)

func paginate[T any](items []T, page, limit int) ([]T, models.PaginationMeta) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = DefaultLimit
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}

	total := len(items)
	if total == 0 {
		return []T{}, models.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      0,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	if page > totalPages && totalPages > 0 {
		return []T{}, models.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    false,
			HasPrev:    page > 1,
		}
	}

	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}

	paginated := items[start:end]

	return paginated, models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

func Paginate(notes []models.Note, page, limit int) models.PaginatedResult {
	items, meta := paginate(notes, page, limit)
	return models.PaginatedResult{
		Notes:      items,
		Pagination: meta,
	}
}

// PaginateSearchResults paginates search results
func PaginateSearchResults(results []models.SearchResult, page, limit int) models.SearchResultsPaginated {
	items, meta := paginate(results, page, limit)
	return models.SearchResultsPaginated{
		Results:    items,
		Pagination: meta,
	}
}
