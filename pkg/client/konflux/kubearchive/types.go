package kubearchive

import (
	"errors"
	"fmt"
	"net/http"
)

type QueryOptions struct {
	LabelSelector string
}

type LogOptions struct {
	Container string
	TailLines int
}

type ListMetadata struct {
	Continue string `json:"continue"`
}

type KubeList[T any] struct {
	Metadata ListMetadata `json:"metadata"`
	Items    []T          `json:"items"`
}

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("KubeArchive API returned %d: %s", e.StatusCode, e.Message)
}

func (e *APIError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	}
	return false
}

type HasUID interface {
	GetUID() string
}
