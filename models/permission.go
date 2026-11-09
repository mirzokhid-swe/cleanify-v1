package models

type Permission struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Scope string `json:"scope"`
	Group string `json:"group"`
}
