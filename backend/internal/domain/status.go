package domain

type Status string

const (
	Active       Status = "active"       // Available for sale, and fully tested
	Discontinued Status = "discontinued" // Not produced anymore, but may exist in stock or replaced by another product.
	Development  Status = "development"  // Available for sale, but not fully tested yet
	Retired      Status = "retired"      // Removed from sale, but still supported
)
