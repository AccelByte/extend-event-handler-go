package server

import (
	"github.com/sony/gobreaker"
)

// StorageWrapper is an interface that represents our storage operations
type StorageWrapper[T any] struct {
	cb     *gobreaker.CircuitBreaker
	client *T
}

func NewStorageWrapper[T any](client *T, cb *gobreaker.CircuitBreaker) *StorageWrapper[T] {
	return &StorageWrapper[T]{client: client, cb: cb}
}

func (c *StorageWrapper[T]) RunWithCb(operation func(*T) (interface{}, error)) (interface{}, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return operation(c.client)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
