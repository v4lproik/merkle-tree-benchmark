package pkg

import (
	"crypto"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"sync"
)

// Hasher is an enum representing a Hash algorithm
type Hasher struct {
	IsSort bool
	Hash   Hash
	Pool   *HashPool
}

type Hash string

const (
	// UNKNOWNHASH is the default value returned for non-supported protocol
	UNKNOWNHASH Hash = "unknown"
	// SHA256 is the identifier for the SHA256 Hash algorithm
	SHA256 Hash = "sha256"
)

var ErrHashNotAllowed = errors.New("Hash<%s> is not recognized")

// IsValid checks if a protocol is valid
func (s Hash) IsValid() bool {
	switch s {
	case SHA256:
		return true
	case UNKNOWNHASH:
		return false
	}
	return false
}

func (s Hash) Hash() crypto.Hash {
	switch s {
	case SHA256:
		return crypto.SHA256
	}
	panic(fmt.Sprintf(ErrHashNotAllowed.Error(), s))
}

func (s Hash) HashFunc() func() hash.Hash {
	switch s {
	case SHA256:
		return sha256.New
	}
	panic(fmt.Sprintf(ErrHashNotAllowed.Error(), s))
}

// ---------------------------------------------------------------------------------------------------------------------

// HashPool is the pool of hashes
type HashPool struct {
	hashFunc sync.Pool
}

// NewHashPool allocates a new pool
func NewHashPool(h crypto.Hash) *HashPool {
	p := &HashPool{}
	p.hashFunc.New = func() interface{} {
		return &hashFunc{Hash: h.New(), pool: &p.hashFunc}
	}
	return p
}

// getHash returns a Hash func instance
func (p *HashPool) getHash() HashCloser {
	return p.hashFunc.Get().(*hashFunc)
}

// hashFunc holds a reference to the pool and the Hash algorithm structure properties
type hashFunc struct {
	hash.Hash
	pool *sync.Pool
}

type HashCloser interface {
	hash.Hash
	Close() error
}

// Close handles the values being put back in the pull once done
func (h *hashFunc) Close() error {
	if h != nil && h.Hash != nil && h.pool != nil {
		h.Hash.Reset()
		h.pool.Put(h)
	}
	return nil
}
