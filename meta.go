package godo

// Meta describes generic information about a response.
type Meta struct {
	Total int `json:"total"`
	Page  int `json:"page,omitempty"`
	Pages int `json:"pages,omitempty"`
}
