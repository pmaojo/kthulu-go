// @kthulu:core
package db

import (
	"database/sql"
	"fmt"
	"strings"

	"backend/internal/repository"
)

// PaginationHelper provides utilities for database pagination
type PaginationHelper struct {
	db *sql.DB
}

// NewPaginationHelper creates a new pagination helper
func NewPaginationHelper(db *sql.DB) *PaginationHelper {
	return &PaginationHelper{db: db}
}

// BuildPaginatedQuery builds a paginated SQL query with sorting
func (h *PaginationHelper) BuildPaginatedQuery(baseQuery string, params repository.PaginationParams, allowedSortFields []string) string {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(baseQuery)

	// Add ORDER BY clause
	if params.SortBy != "" && h.isValidSortField(params.SortBy, allowedSortFields) {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", params.SortBy, strings.ToUpper(params.SortDir)))
	} else if len(allowedSortFields) > 0 {
		// Default to first allowed field
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", allowedSortFields[0], strings.ToUpper(params.SortDir)))
	}

	// Add LIMIT and OFFSET
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", params.PageSize, params.CalculateOffset()))

	return queryBuilder.String()
}

// BuildCountQuery builds a count query from a base query
func (h *PaginationHelper) BuildCountQuery(baseQuery string) string {
	// Remove ORDER BY clause if present (not needed for count)
	if orderByIndex := strings.Index(strings.ToUpper(baseQuery), " ORDER BY"); orderByIndex != -1 {
		baseQuery = baseQuery[:orderByIndex]
	}

	// Wrap in COUNT query
	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_query", baseQuery)
}

// ExecutePaginatedQuery executes a paginated query and returns results with metadata
func (h *PaginationHelper) ExecutePaginatedQuery(
	baseQuery string,
	countQuery string,
	params repository.PaginationParams,
	allowedSortFields []string,
	scanFunc func(*sql.Rows) (interface{}, error),
	args ...interface{},
) (interface{}, int64, error) {
	// Get total count
	var total int64
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build paginated query
	paginatedQuery := h.BuildPaginatedQuery(baseQuery, params, allowedSortFields)

	// Execute paginated query
	rows, err := h.db.Query(paginatedQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute paginated query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []interface{}
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return results, total, nil
}

// isValidSortField checks if a sort field is in the allowed list
func (h *PaginationHelper) isValidSortField(field string, allowedFields []string) bool {
	for _, allowed := range allowedFields {
		if field == allowed {
			return true
		}
	}
	return false
}

// BuildSearchQuery builds a search query with LIKE conditions
func (h *PaginationHelper) BuildSearchQuery(baseQuery string, searchFields []string, searchTerm string) (string, []interface{}) {
	if searchTerm == "" || len(searchFields) == 0 {
		return baseQuery, nil
	}

	var conditions []string
	var args []interface{}
	searchPattern := "%" + searchTerm + "%"

	for _, field := range searchFields {
		conditions = append(conditions, fmt.Sprintf("%s ILIKE ?", field))
		args = append(args, searchPattern)
	}

	searchCondition := "(" + strings.Join(conditions, " OR ") + ")"

	// Add WHERE or AND depending on whether base query already has WHERE
	if strings.Contains(strings.ToUpper(baseQuery), " WHERE ") {
		baseQuery += " AND " + searchCondition
	} else {
		baseQuery += " WHERE " + searchCondition
	}

	return baseQuery, args
}

// BuildFilterQuery builds a query with additional filters
func (h *PaginationHelper) BuildFilterQuery(baseQuery string, filters map[string]interface{}) (string, []interface{}) {
	if len(filters) == 0 {
		return baseQuery, nil
	}

	var conditions []string
	var args []interface{}

	for field, value := range filters {
		if value != nil {
			conditions = append(conditions, fmt.Sprintf("%s = ?", field))
			args = append(args, value)
		}
	}

	if len(conditions) == 0 {
		return baseQuery, nil
	}

	filterCondition := strings.Join(conditions, " AND ")

	// Add WHERE or AND depending on whether base query already has WHERE
	if strings.Contains(strings.ToUpper(baseQuery), " WHERE ") {
		baseQuery += " AND " + filterCondition
	} else {
		baseQuery += " WHERE " + filterCondition
	}

	return baseQuery, args
}

// CombineArgs combines multiple argument slices
func (h *PaginationHelper) CombineArgs(argSlices ...[]interface{}) []interface{} {
	var combined []interface{}
	for _, args := range argSlices {
		combined = append(combined, args...)
	}
	return combined
}
