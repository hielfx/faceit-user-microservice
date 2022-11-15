package pagination

// SortOrder - sort asc or desc
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// Valid -  returns true if the SortOrder is valid
func (so SortOrder) Valid() bool {
	return so == SortOrderAsc || so == SortOrderDesc
}

// IsAsc - returns true if SortOrder is asc
func (so SortOrder) IsAsc() bool {
	return so == SortOrderAsc
}

// IsDesc - returns true if SortOrder is desc
func (so SortOrder) IsDesc() bool {
	return so == SortOrderDesc
}
