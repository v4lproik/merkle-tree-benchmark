package pkg

import (
	"bytes"
	"fmt"
)

// Node represents a node within the tree
// a node can be defined as a leaf or a parent node - calculated from two leaves or two child nodes
type Node struct {
	Parent   *Node
	Left     *Node
	Right    *Node
	isOrphan bool
	Hash     []byte
	Data     Data
}

func NewLeaf(p *Hasher, d Data) (*Node, error) {
	return newLeaf(p, d, false)
}

func NewOrphanLeaf(p *Hasher, d Data) (*Node, error) {
	return newLeaf(p, d, true)
}

func NewParentNode(p *Hasher, left, right *Node) (*Node, error) {
	if p.Pool == nil {
		hf := p.Hash.HashFunc()()
		if _, err := hf.Write(concat(false, p.IsSort, left.Hash, right.Hash)); err != nil {
			return nil, fmt.Errorf("hf.Write(concat(%x,%x)): %w", left.Hash, right.Hash, err)
		}
		return &Node{
			Left:  left,
			Right: right,
			Hash:  hf.Sum(nil),
		}, nil
	}

	h := p.Pool.getHash()
	defer h.Close()

	if _, err := h.Write(concat(true, p.IsSort, left.Hash, right.Hash)); err != nil {
		return nil, fmt.Errorf("hf.Write(concat(%x,%x)): %w", left.Hash, right.Hash, err)
	}

	return &Node{
		Left:  left,
		Right: right,
		Hash:  h.Sum(nil),
	}, nil
}

func newLeaf(p *Hasher, d Data, isPadding bool) (*Node, error) {
	var (
		err error
		b   []byte
	)

	if b, err = d.Hash(p); err != nil {
		return nil, fmt.Errorf("d.Hasher(): data<%s>: %w", d, err)
	}

	return &Node{
		isOrphan: isPadding,
		Hash:     b,
		Data:     d,
	}, nil
}

func (n *Node) isLeaf() bool {
	return n.Left == nil && n.Right == nil
}

type NodeSorter struct {
	nodes []*Node
}

func (n NodeSorter) Len() int {
	return len(n.nodes)
}

func (n NodeSorter) Swap(i, j int) {
	n.nodes[i], n.nodes[j] = n.nodes[j], n.nodes[i]
}

func (n NodeSorter) Less(i, j int) bool {
	return bytes.Compare(n.nodes[i].Hash, n.nodes[j].Hash) == -1
}
