package dht

import (
	"fmt"
	"math/big"
)

const (
	m      = 161 // 1-base indexing
	suSize = 3
)

// Node of a virtual machine-- each resource is in Data
type Node struct {
	// HOst + ":" + Port = address of node(server), using n.addr() to get
	Host        string
	Port        string
	Id          *big.Int           // hash(addr)
	Successor   [suSize + 1]string // do at 1, 2, 3; using 0 to do Node, so that can simplify
	Predecessor string

	Data        map[string]string
	FingerTable [m]string
	Next        int
}

type KVP struct {
	K, V string //key, value
}

func NewNode(_port string) *Node {
	_host := GetAddress()

	return &Node{
		Host: _host,
		Port: _port,
		Id:   Hash(_host + ":" + _port),
		Data: make(map[string]string),
	}
}

func (n *Node) Addr() string {
	return n.Host + ":" + n.Port
}

//init predecessor and successor
func (n *Node) create() {
	n.Predecessor = n.Addr() // or "", can be stablized later
	for i, _ := range n.Successor {
		n.Successor[i] = n.Addr()
	}

	//go----
}

//init node's info
func (n *Node) join(address string) error {
	n.Predecessor = "" // have to be stablized later

	n.Successor[0] = n.Addr()
	for i := 1; i <= suSize; i++ {
		addr, err := RPCFindSuccessor(address, Hash(n.Successor[i-1]))
		if err != nil {
			fmt.Printf("node join-findsuccessor %v\n", err)
			return err
		}
		n.Successor[i] = addr
	}

	return nil
}

func (n *Node) Ping(request int, response *int) error {
	//do something with request?? no need
	fmt.Println("Get request for Ping: ", request)
	*response = len(n.Data)
	return nil
}

//If the id I'm looking for falls between me and mySuccessor
//Then the data for this id will be found on mySuccessor
// ------------>>>will do finger table later
// ask node n to find the successor of id
// or a better node to continue the search with
func (n *Node) FindSuccessor(id *big.Int, successor *string) error {
	n.Successor[0] = n.Addr()
	for i := 0; i < suSize; i++ {
		fmt.Println("successor are: ", n.Successor[i], n.Successor[i+1])
		if InclusiveBetween(id, Hash(n.Successor[i]), Hash(n.Successor[i])) { //id ∈ (n, successor]
			*successor = n.Successor[i+1]
			fmt.Println("successor find: ", n.Successor[i+1])
			return nil
		}
	}

	var err error
	*successor, err = RPCFindSuccessor(n.Successor[3], id)
	return err
	// return false, closest_preceding_node(id);

	// search the local table for the highest predecessor of id
	// n.closest_preceding_node(id)
	// // skip this loop if you do not have finger tables implemented yet
	// for i = m downto 1
	//     if (finger[i] ∈ (n,id])
	//         return finger[i];
	// return successor;

	// find the successor of id
	// find(id, start)
	//     found, nextNode = false, start;
	//     i = 0
	//     while not found and i < maxSteps
	//         found, nextNode = nextNode.find_successor(id);
	//         i += 1
	//     if found
	//         return nextNode;
	//     else
	//         report error;
}

// func (n *Node) Notify(addr string, response *bool) error           { return nil }
// func (n *Node) GetPredecessor(none bool, addr *string) error       { return nil }

func (n *Node) Put(args KVP, success *bool) error {
	n.Data[args.K] = args.V
	*success = true
	return nil
}
func (n *Node) Get(key string, response *string) error {
	*response = n.Data[key]
	return nil
}
func (n *Node) Del(key string, success *bool) error {
	if _, ok := n.Data[key]; ok {
		delete(n.Data, key)
		*success = true
		return nil
	} else {
		*success = false
		return nil //ok func, but can't find key
	}
}

// func (n *Node) stabalize()             {}
// func (n *Node) stabalizeOften()        {}
// func (n *Node) checkPredecessor()      {}
// func (n *Node) checkPredecessorOften() {}
// func (n *Node) fixFingerTable()        {}
// func (n *Node) fixFingerTableOften()   {}

//this is a test function to check if rpc is working
func (n *Node) Test(curmsg string, size *int) error {
	n.Data[curmsg] = ""

	*size = 0 //cnt again
	for msg, _ := range n.Data {
		*size++
		fmt.Println(msg) // cur do at server, so get msg printed at server
	}
	return nil
}
