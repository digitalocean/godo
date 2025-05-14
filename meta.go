package godo

// Meta describes generic information about a response.
type Meta struct {
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}
