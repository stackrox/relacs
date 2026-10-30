package kubearchive

import "fmt"

type QueryOptions struct {
	LabelSelector string
	Limit         int
}

type LogOptions struct {
	Container string
	TailLines int
}

type KubeList[T any] struct {
	Items []T `json:"items"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("KubeArchive API returned %d: %s", e.StatusCode, e.Message)
}

type HasUID interface {
	GetUID() string
}
