package shared

import "strconv"

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (p Pagination) Normalized() Pagination {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return p
}

func (p Pagination) Offset() int {
	p = p.Normalized()
	return (p.Page - 1) * p.Limit
}

func IntFromString(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
