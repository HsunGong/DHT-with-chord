package dht

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"
)

const (
	m           = 160 // 0-base indexing
	suSize      = 3
	refreshTime = 500 * time.Millisecond
)

// Node of a virtual machine-- each resource is in Data
type Node struct {
	// HOst + ":" + Port = address of node(server), using n.addr() to get
	Host string
	Port string
	// Id             *big.Int // hash(addr)
	Address        string
	SuccessorTable [suSize + 1]string // do at 1, 2, 3; using 0 to do successor, so that can simplify
	successor      string
	Predecessor    string

	Data        map[string]string
	FingerTable [m + 1]string
	next        int

	Listening bool
	debug     bool // if true, is debugging
}

type KVP struct {
	K, V string //key, value
}

func NewNode(_port string, _debug bool) *Node {
	_host := GetAddress()

	p := &Node{
		Host: _host,
		Port: _port,
		// Id:    Hash(fmt.Sprintf("%v:%v", _host, _port)),
		Address: fmt.Sprintf("%v:%v", _host, _port),
		Data:    make(map[string]string),
		debug:   _debug,
		// FingerTable: make([]string, 0, m),
		next: 0,
	}
	return p
}

//init predecessor and successor
func (n *Node) create() {
	n.Predecessor = "" //Addr(n) // or "", can be stablized later
	n.successor = n.Address
	for i, _ := range n.SuccessorTable {
		n.SuccessorTable[i] = n.Address
	}

	go n.checkSurvival()
	go n.checkPredecessorPeriod()
	go n.stabilizePeriod()
	go n.fixFingerTablePeriod()
}

//init node's info
func (n *Node) join(address string) error {
	n.Predecessor = "" // have to be stablized later

	addr, err := RPCFindSuccessor(address, Hash(n.Address))
	if err != nil {
		fmt.Printf("node join-findsuccessor %v\n", err)
		return err
	}
	n.successor = addr
	if err := n.fixSuccessorTable(); err != nil {
		return err
	}

	//notify the successor[1]
	fmt.Println("notify")
	if err := RPCNotify(n.successor, n.Address); err != nil {
		fmt.Println("Join but Notify err")
	}
	n.stabilize()

	return nil
}

func (n *Node) CopySuccessor(a bool, table *[suSize + 1]string) error {
	for i := 0; i <= suSize; i++ {
		(*table)[i] = n.SuccessorTable[i]
	}
	return nil
}

func (n *Node) fixSuccessorTable() error {
	var a bool
	if err := Call(n.successor, "Node.CopySuccessor", a, &n.SuccessorTable); err != nil {
		return err
	}

	for i := 0; i < suSize; i++ {
		n.SuccessorTable[i+1] = n.SuccessorTable[i]
	}
	n.SuccessorTable[0] = n.successor
	return nil
}

// search the local table for the highest predecessor of id
// n.closest_preceding_node(id)
//if empty, return "", left to let
// find finger[i] ∈ (n, id)
func (n *Node) findFingerTable(id *big.Int) string {
	if len(n.FingerTable) == 0 {
		return ""
	}

	slice := n.FingerTable[:]
	for i := len(slice); i >= 0; i-- { ///??????
		if !ExclusiveBetween(Hash(n.FingerTable[i]), Hash(n.Address), id) {
			// the return should be i - 1
			// what if i == 0?, impossible for successor have to be included
			if i == 0 {
				log.Panicln("FingerTable doesnt fit with successor")
			}
			return n.FingerTable[i-1]
		}
	}

	// all of them is included by n.Id and id, so as said should return n
	//however incase loop forever, return "", let the function to handle
	return ""

}

//If the id I'm looking for falls between me and mySuccessor
//Then the data for this id will be found on mySuccessor
// ------------>>>will do finger table later
// ask node n to find the successor of id
// or a better node to continue the search with
//!!!!!maybe can repete with some node(in fixfinger)
func (n *Node) FindSuccessor(id *big.Int, successor *string) error {
	// fmt.Println(Hash(n.Address), id, n.Address)
	if InclusiveBetween(id, Hash(n.Address), Hash(n.successor)) {
		*successor = n.successor

		return nil
	}

	for i := 1; i <= suSize; i++ {
		// fmt.Println("circle")
		if InclusiveBetween(id, Hash(n.SuccessorTable[i-1]), Hash(n.SuccessorTable[i])) { //id ∈ (n, successor]
			*successor = n.SuccessorTable[i]
			// fmt.Println(id, " successor find: ", n.SuccessorTable[i])
			return nil
		}
	}

	//wont loop if successor is exists

	nextAddr := n.findFingerTable(id)
	if nextAddr == "" {
		return errors.New("findFingerTable Err")
	} else {
		var err error
		*successor, err = RPCFindSuccessor(nextAddr, id)
		return err
	}

}

func (n *Node) GetPredecessor(none bool, addr *string) error {
	*addr = n.Predecessor
	if n.Predecessor == "" {
		return errors.New("Predecessor is Empty")
	} else {
		return nil
	}
}

//address is the node that should be awared
//address maybe the n's Predecessor
func (n *Node) Notify(new_predecessor string, response *bool) error {
	if n.Predecessor == "" || ExclusiveBetween(Hash(new_predecessor), Hash(n.Predecessor), Hash(n.Address)) {
		if n.debug {
			fmt.Printf("%s Notify %s\n", n.Address, new_predecessor)
		}
		n.Predecessor = new_predecessor
		*response = true
	}
	*response = false
	return nil
}

//check whther predecessor is failed
func (n *Node) checkPredecessor() {
	response := 0
	if err := Call(n.Predecessor, "Node.Ping", 101, &response); err != nil {
		n.Predecessor = ""
	}
}

//call periodly, refresh.
//n.next is the index of finger to fix
func (n *Node) fixFingerTable() {
	// n.next += 1
	// if n.next >= m {
	// 	n.next = 0
	// }

	id := fingerEntry(Hash(n.Address), n.next)
	var response string
	if err := n.FindSuccessor(id, &response); err != nil || response == "" {
		fmt.Printf("fixFingertable err at: %v\n", err)
		return
	}

	// fmt.Printf("cur: %d and %d, resp: %d\n", n.Id, Hash(Addr(n)), Hash(response))
	for InclusiveBetween(id, Hash(n.Address), Hash(response)) {
		n.FingerTable[n.next] = response
		// fmt.Printf("fixed [%d] = %s\n", n.next, response)

		n.next += 1
		if n.next >= m {
			n.next = 0
			break
		}
		id = fingerEntry(Hash(n.Address), n.next)
	}
}

//Call periodly, verfiy succcessor, tell the successor of n
//only check successor 1
func (n *Node) stabilize() {
	pre_i, err := RPCGetPredecessor(n.successor)

	// FIND A PREDECESSOR
	// a-->c(in a), c-->b(in c), ==> a<-->b<-->c
	// ?? belong (n, successor) ??
	if err == nil {
		if pre_i != n.Address { // ?????
			//if ExclusiveBetween(Hash(p), Hash(n.Successor[i-1], Hash(n.Successor[i]) {
			fmt.Printf("change predecessor: %s --> [%s] --> %s\n", n.successor, pre_i, n.Address)

			n.successor = pre_i
			if err := n.fixSuccessorTable(); err != nil {
				n.successor = n.SuccessorTable[0]
				RPCNotify(n.successor, n.Address)
			}
		}
		goto SKIP
	} else { // no this successor[i] or its predecessor
		for i := 1; i <= suSize; i++ {
			n.successor = n.SuccessorTable[i]
			pre_i, err := RPCGetPredecessor(n.successor)
			if err != nil {
				continue
			}

			if pre_i != n.Address { // ?????
				fmt.Printf("change predecessor: %s --> [%s] --> %s\n", n.successor, pre_i, n.Address)
				n.successor = pre_i
				if err := n.fixSuccessorTable(); err != nil {
					n.successor = n.SuccessorTable[0]
					// RPCNotify(n.successor, Addr(n))
				}
			}
			goto SKIP
		}
	}

	log.Println("Successor List Failed")

SKIP:
	if err = RPCNotify(n.successor, n.Address); err != nil {
		fmt.Printf("Stabilize with Notfy err: %v", err)
	}
}

func (n *Node) checkSurvival() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			// fmt.Println("check stablize")
			response := 0
			if err := Call(n.Address, "Node.Ping", 51, &response); err != nil {
				n.Listening = false
				return
			}
		}
	}

}

//using 1 func-- 1 tick strategy, can not sync with(using frequency maybe different)
func (n *Node) stabilizePeriod() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			// fmt.Println("check stablize")
			n.stabilize()
		}
	}

}
func (n *Node) checkPredecessorPeriod() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			// fmt.Println("check Pre")
			n.checkPredecessor()
		}
	}
}
func (n *Node) fixFingerTablePeriod() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			n.fixFingerTable()
		}
	}
}

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

func (n *Node) Adapt(datas map[string]string, success *bool) error {
	for k, v := range datas {
		n.Data[k] = v
	}
	*success = true
	return nil
}

func (n *Node) Ping(request int, response *int) error {
	//do something with request?? no need
	*response = len(n.Data)
	return nil
}
