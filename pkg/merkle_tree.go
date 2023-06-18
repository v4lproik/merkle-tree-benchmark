package pkg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"sort"
)

// MerkleTree is the data structure representing a tree
// it contains the root which is a concatenation of all the hashes from all the nodes of the tree
type MerkleTree struct {
	Root   *Node
	Leaves []*Node
	MerkleTreeConfig
}

// MerkleTreeConfig is the configuration that represents the options used to build / verify the tree
type MerkleTreeConfig struct {
	Hasher       *Hasher
	MaxGoroutine uint32
	isSort       bool
}

// MerkleTreeBuilder allows use to pass the configuration from the cli before building a tree
type MerkleTreeBuilder struct {
	config *MerkleTreeConfig
}

var (
	ErrMerkleTreeConfigIsNil                = errors.New("the merkle tree configWithHashPool cannot be nil")
	ErrMerkleTreeConfigHasherIsNil          = errors.New("the merkle tree configWithHashPool hasher cannot be nil")
	ErrMerkleTreeConfigMaxGoroutineIsEqZero = errors.New("the merkle tree configWithHashPool max goroutine cannot be equal to 0")
	ErrMerkleTreeDataIsNilOrEmpty           = errors.New("the merkle tree data cannot be nil or empty")
)

func NewMerkleTreeBuilder() *MerkleTreeBuilder {
	return &MerkleTreeBuilder{config: &MerkleTreeConfig{}}
}

func (b *MerkleTreeBuilder) WithHasher(hasher *Hasher) *MerkleTreeBuilder {
	b.config.Hasher = hasher
	return b
}

func (b *MerkleTreeBuilder) WithMaxGoroutine(maxGoroutine uint32) *MerkleTreeBuilder {
	b.config.MaxGoroutine = maxGoroutine
	return b
}

// Build builds the tree with the data passed parameter
// we allow the passage of a context in order to be able to stop the execution from the caller if needed
func (b *MerkleTreeBuilder) Build(ctx context.Context, data []Data) (*MerkleTree, error) {
	var (
		mt        *MerkleTree
		leafNodes []*Node

		err error
	)

	// conditions to generate a tree
	if b.config == nil {
		return mt, ErrMerkleTreeConfigIsNil
	}

	if b.config.Hasher == nil {
		return mt, ErrMerkleTreeConfigHasherIsNil
	}

	if b.config.MaxGoroutine == 0 {
		return mt, ErrMerkleTreeConfigMaxGoroutineIsEqZero
	}

	if len(data) == 0 {
		return mt, ErrMerkleTreeDataIsNilOrEmpty
	}

	// init merkle tree object
	mt = &MerkleTree{
		Root:             nil,
		Leaves:           nil,
		MerkleTreeConfig: *b.config,
	}

	// build tree
	if leafNodes, err = mt.generateLeafNodes(ctx, data); err != nil {
		return mt, fmt.Errorf("mt.generateLeafNodes(data): %w", err)
	}

	if mt.Root, err = mt.generateParentNodes(ctx, leafNodes); err != nil {
		return mt, fmt.Errorf("mt.generateParentNodes(): %w", err)
	}

	mt.Leaves = leafNodes

	return mt, nil
}

// generateLeafNodes generates an array of Nodes that represents the leaves placed at the bottom of the tree
// it handles the case where there's an uneven nb of leaves in the tree
func (mt *MerkleTree) generateLeafNodes(ctx context.Context, data []Data) ([]*Node, error) {
	if len(data) == 0 {
		return nil, ErrMerkleTreeDataIsNilOrEmpty
	}

	var (
		leaves       []*Node
		isUnevenData = len(data)%2 == 1
	)

	// generate bottom leaves
	// handle use case where there's an uneven nb of leaves (it always goes by pair)
	// perf: better to use make here than using append which doubles the array increasing memory pressure
	// use allocation here to avoid handling concurrent writes with a lock
	if isUnevenData {
		leaves = make([]*Node, len(data)+1)
	} else {
		leaves = make([]*Node, len(data))
	}

	// create leaves
	errs, _ := errgroup.WithContext(ctx)
	errs.SetLimit(int(mt.MerkleTreeConfig.MaxGoroutine))
	for _i := 0; _i < len(data); _i++ {
		// i can change in the below go routine, allocates a local scope via i
		i := _i

		errs.Go(func() error {
			leaf, err := NewLeaf(mt.Hasher, data[i])
			if err != nil {
				return fmt.Errorf("NewLeaf(data[%d]): %w", i, err)
			}
			log.Debugf("new leaf: val<%s>=Hash<%x>", leaf.Data, leaf.Hash)
			leaves[i] = leaf
			return nil
		})
	}

	// wait for all the go routines to be done
	if err := errs.Wait(); err != nil {
		return nil, err
	}

	// create last leaf - duplicate the last leaf to have a even number of leaves in the tree
	if isUnevenData {
		d := data[len(data)-1]
		leaf, err := NewOrphanLeaf(mt.Hasher, d)
		if err != nil {
			return nil, err
		}
		leaves[len(data)] = leaf
	}

	if mt.Hasher.IsSort {
		sort.Sort(NodeSorter{nodes: leaves})
	}

	return leaves, nil
}

// generateParentNodes generates a parent node by pairing two nodes together
func (mt *MerkleTree) generateParentNodes(ctx context.Context, leafNodes []*Node) (*Node, error) {
	if len(leafNodes) == 0 {
		return nil, ErrMerkleTreeDataIsNilOrEmpty
	}

	var (
		nodes   []*Node
		counter int

		isUnevenNode = len(leafNodes)%2 == 1
	)

	// generate a parent node from 2 nodes' hashes
	// once again calculate the array and use allocation instead of append which is prone to error with goroutines
	if isUnevenNode {
		nodes = make([]*Node, len(leafNodes)/2+1)
	} else {
		nodes = make([]*Node, len(leafNodes)/2)
	}

	errs, _ := errgroup.WithContext(ctx)
	errs.SetLimit(int(mt.MerkleTreeConfig.MaxGoroutine))
	for _i := 0; _i < len(leafNodes); _i += 2 {
		left, right := _i, _i+1
		c := counter

		errs.Go(func() error {
			// if orphan node, we need to Hash it twice to respect the binary property of the tree
			if left+1 == len(leafNodes) {
				right = left
			}

			// generate parent node Hash
			node, err := NewParentNode(mt.Hasher, leafNodes[left], leafNodes[right])
			if err != nil {
				return fmt.Errorf("NewParentNode(): %w", err)
			}
			log.Debugf("new parent: val<%x,%x>=Hash<%x>", leafNodes[left].Hash, leafNodes[right].Hash, node.Hash)

			// refer each leaf to its freshly generated parent node
			leafNodes[left].Parent = node
			leafNodes[right].Parent = node

			nodes[c] = node

			return nil
		})
		counter++
	}

	// wait for all the go routines to be done
	if err := errs.Wait(); err != nil {
		return nil, err
	}

	// we have calculated the last pair available, in sum, the tree root
	if len(leafNodes) == 2 {
		return leafNodes[len(leafNodes)-1].Parent, nil
	}

	// otherwise let's keep it calculating the parent nodes up to the merkle tree root
	return mt.generateParentNodes(ctx, nodes)
}

// Verify verifies if a leaf containing the data passed in parameter is present in the tree
// it calculates the hash of all the parents nodes all the way to the tree root
// if one hash is different than its parent's, false is returned
func (mt *MerkleTree) Verify(context context.Context, data Data) (bool, error) {
	if mt.Leaves == nil || len(mt.Leaves) == 0 {
		log.Warn("tree is empty or doesn't contain any nodes")
		return false, nil
	}

	// calculate the data Hash
	hash, err := data.Hash(mt.Hasher)
	if err != nil {
		return false, fmt.Errorf("data.Hasher(): %w", err)
	}

	for _, leaf := range mt.Leaves {
		if !bytes.Equal(leaf.Hash, hash) {
			continue
		}

		currentParent := leaf.Parent
		for currentParent != nil {
			var (
				leftNodeHash, rightNodeHash []byte
			)

			if leftNodeHash, err = mt.computeNodeHash(currentParent.Left); err != nil {
				return false, fmt.Errorf("mt.computeNodeHash(currentParent.Left): %w", err)
			}

			if rightNodeHash, err = mt.computeNodeHash(currentParent.Right); err != nil {
				return false, fmt.Errorf("mt.computeNodeHash(currentParent.Right): %w", err)
			}

			if mt.Hasher.Pool == nil {
				hf := mt.Hasher.Hash.HashFunc()()
				if _, err = hf.Write(concat(false, mt.Hasher.IsSort, leftNodeHash, rightNodeHash)); err != nil {
					return false, fmt.Errorf("hf.Write(concat(%x,%x)): %w", leftNodeHash, rightNodeHash, err)
				}

				if !bytes.Equal(hf.Sum(nil), currentParent.Hash) {
					return false, nil
				}

				currentParent = currentParent.Parent
				continue
			}

			hf := mt.Hasher.Pool.getHash()
			defer hf.Close()

			if _, err = hf.Write(concat(true, mt.Hasher.IsSort, leftNodeHash, rightNodeHash)); err != nil {
				return false, fmt.Errorf("hf.Write(concat(%x,%x)): %w", leftNodeHash, rightNodeHash, err)
			}

			if !bytes.Equal(hf.Sum(nil), currentParent.Hash) {
				return false, nil
			}

			currentParent = currentParent.Parent
		}
		return true, nil
	}
	return false, nil
}

// computeNodeHash firstly determines if the node is a leaf or a parent node
// a leaf is only calculate such as H(data) whereas a parent node is calculated such as H(Hl(data)+Hr(data))
func (mt *MerkleTree) computeNodeHash(n *Node) ([]byte, error) {
	if n.isLeaf() {
		return n.Data.Hash(mt.Hasher)
	}
	if mt.Hasher.Pool == nil {
		hf := mt.Hasher.Hash.HashFunc()()
		if _, err := hf.Write(concat(false, mt.Hasher.IsSort, n.Left.Hash, n.Right.Hash)); err != nil {
			return nil, fmt.Errorf("hf.Write(concat(%x,%x)): %w", n.Left.Hash, n.Right.Hash, err)
		}
		return hf.Sum(nil), nil
	}

	h := mt.Hasher.Pool.getHash()
	defer h.Close()

	if _, err := h.Write(concat(true, mt.Hasher.IsSort, n.Left.Hash, n.Right.Hash)); err != nil {
		return nil, fmt.Errorf("hf.Write(concat(%x,%x)): %w", n.Left.Hash, n.Right.Hash, err)
	}
	return h.Sum(nil), nil
}
