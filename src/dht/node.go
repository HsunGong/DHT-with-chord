package dht

import (
	"math/big"
)

const m = 161 // 1-base indexing

// Node of a virtual machine-- each resource is in Data
type Node struct {
	addr        string //addres
	port        string
	id          *big.Int // hash(addr)
	Successor   [3]string
	Predecessor string

	Data        map[string]string
	fingerTable [m]string
	next        int
}

type KVP struct {
	k, v string //key, value
}

func NewNode(port string) *Node {}

func (n *Node) GetAddr() string                                    {}
func (n *Node) GetPing(first int, second *int) error               {}
func (n *Node) Put(args KVP, success *bool) error                  {}
func (n *Node) Get(key string, response *string) error             {}
func (n *Node) Del(key string, success *bool) error                {}
func (n *Node) FindSuccessor(id *big.Int, successor *string) error {}
func (n *Node) Notify(addr string, response *bool) error           {}
func (n *Node) GetPredecessor(none bool, addr *string) error       {}
func (n *Node) join(addr string)                                   {}
func (n *Node) stabalize()                                         {}
func (n *Node) stabalizeOften()                                    {}
func (n *Node) checkPredecessor()                                  {}
func (n *Node) checkPredecessorOften()                             {}
func (n *Node) fixFingerTable()                                    {}
func (n *Node) fixFingerTableOften()                               {}
func (n *Node) create()                                            {}
