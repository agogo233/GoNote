package models

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
    HasNext    bool `json:"has_next"`
    HasPrev    bool `json:"has_prev"`
}

// PaginatedResult represents a paginated list of notes
type PaginatedResult struct {
    Notes     []Note        `json:"notes"`
    Pagination PaginationMeta `json:"pagination"`
}
