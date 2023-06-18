package pkg

import (
	"fmt"
)

// Data is the interface representing a data structure containing a piece of data and that can be hashed
type Data interface {
	Hash(h *Hasher) ([]byte, error)
	String() string
}

// ---------------------------------------------------------------------------------------------------------------------

// StringData represents a data of type string
type StringData struct {
	Value string
}

func (s StringData) Hash(h *Hasher) ([]byte, error) {
	if h.Pool == nil {
		hf := h.Hash.HashFunc()()
		if _, err := hf.Write([]byte(s.Value)); err != nil {
			return nil, fmt.Errorf("hf.Write(%s): %w", s.Value, err)
		}
		return hf.Sum(nil), nil
	}

	hf := h.Pool.getHash()
	defer hf.Close()

	if _, err := hf.Write([]byte(s.Value)); err != nil {
		return nil, fmt.Errorf("hf.Write(%s): %w", s.Value, err)
	}
	return hf.Sum(nil), nil
}

func (s StringData) String() string {
	return s.Value
}
