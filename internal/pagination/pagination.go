package pagination

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const (
	DefaultSize = 10
	FirstPage   = 1
)

// PaginationOptions struct with the pagination data
type PaginationOptions struct {
	Size int `json:"size" query:"size"`
	Page int `json:"page" query:"page"`
	// OrderBy   string    `json:"orderBy"`   // OrderBy property to order by
	// SortOrder SortOrder `json:"sortOrdet"` // Sort order (asc or desc)
}

// Paginated is used as response and has more information
type Paginated struct {
	TotalCount  int64 `json:"totalCount"` //TotalCount is the number of elements that has the db
	TotalPages  int64 `json:"totalPages"` //TotalPages is the number of pages based on the total count
	CurrentPage int   `json:"currentPage"`
	Size        int   `json:"size"`
	HasMore     bool  `json:"hasMore"`
}

// FromEchoContext - Returns a new PaginationOptions from the Echo context
func FromEchoContext(c echo.Context) (*PaginationOptions, error) {
	var page, size int
	var err error
	queryPage := c.QueryParam("page")

	if queryPage == "" {
		page = FirstPage
	} else {
		page, err = strconv.Atoi(queryPage)
		if err != nil {
			logrus.Errorf("Error in pagination.FromEchoContext -> error retrieving page: %s", err)
			return nil, err
		}
		if page < 0 {
			page = FirstPage
		}
	}

	querySize := c.QueryParam("size")
	if querySize == "" {
		size = DefaultSize
	} else {
		size, err = strconv.Atoi(querySize)
		if err != nil {
			logrus.Errorf("Error in pagination.FromEchoContext -> error retrieving size: %s", err)
			return nil, err
		}
		if size < 0 {
			size = DefaultSize
		}
	}

	// orderBy := c.QueryParam("orderBy")

	sortOrder := SortOrder(c.QueryParam("sort"))
	if !sortOrder.Valid() {
		sortOrder = SortOrderDesc
	}

	return &PaginationOptions{
		Page: page,
		Size: size,
		// OrderBy:   orderBy,
		// SortOrder: sortOrder,
	}, nil
}
