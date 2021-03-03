package chain_node

import "github.com/Xarepo/msc-container-migration/internal/dump"

type ChainNode struct {
	el     *dump.Dump
	prev   *ChainNode
	synced bool
}

func New(el *dump.Dump, prev *ChainNode, synced bool) *ChainNode {
	return &ChainNode{el, prev, synced}
}

func (node ChainNode) Dump() *dump.Dump {
	return node.el
}

func (node *ChainNode) SetSynced() {
	node.synced = true
}

func (node ChainNode) IsSynced() bool {
	return node.synced
}

func (node *ChainNode) GetPrev() *ChainNode {
	return node.prev
}
