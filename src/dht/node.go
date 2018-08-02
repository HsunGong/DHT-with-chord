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
	refreshTime = 0.5 * time.Second
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
	FingerTable []string
	next        int

	debug bool // if true, is debugging
}

type KVP struct {
	K, V string //key, value
}

func NewNode(_port string, _debug bool) *Node {
	_host := GetAddress()

	return &Node{
		Host:        _host,
		Port:        _port,
		Id:          Hash(_host + ":" + _port),
		Data:        make(map[string]string),
		debug:       _debug,
		FingerTable: make([]string, 0, m),
		next:        0,
	}
}

func Addr(n *Node) string {
	return n.Host + ":" + n.Port
}

//init predecessor and successor
func (n *Node) create() {
	n.Predecessor = Addr(n) // or "", can be stablized later
	for i, _ := range n.Successor {
		n.Successor[i] = Addr(n)
	}

	//go----
	go n.stabalizePeriod()
	go n.checkPredecessorPeriod()
	go n.fixFingerTablePeriod()
}

//init node's info
func (n *Node) join(address string) error {
	n.Predecessor = "" // have to be stablized later

	n.Successor[0] = Addr(n)
	for i := 1; i <= suSize; i++ {
		addr, err := RPCFindSuccessor(address, Hash(n.Successor[i-1]))
		if err != nil {
			fmt.Printf("node join-findsuccessor %v\n", err)
			return err
		}
		n.Successor[i] = addr
		if n.debug {
			fmt.Printf("successor[%d] are: %v\n", i, n.Successor[i])
		}
	}

	return nil
}

func (n *Node) Ping(request int, response *int) error {
	//do something with request?? no need
	fmt.Println("Get request for Ping: ", request)
	*response = len(n.Data)
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

	for i := len(n.FingerTable); i >= 0; i-- {
		if n.debug {
			fmt.Println(ExclusiveBetween(Hash(n.FingerTable[i]), n.Id, id))
		}

		if !ExclusiveBetween(Hash(n.FingerTable[i]), n.Id, id) {
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
	n.Successor[0] = Addr(n)
	for i := 1; i <= suSize; i++ {
		if n.debug {
			fmt.Printf("successor[%d] are: (%v) %v\n", i, n.Successor[i-1], n.Successor[i])
		}
		if InclusiveBetween(id, Hash(n.Successor[i-1]), Hash(n.Successor[i])) { //id ∈ (n, successor]
			*successor = n.Successor[i]
			if n.debug {
				fmt.Println("successor find: ", n.Successor[i])
			}
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
func (n *Node) Notify(address string, response *bool) error {
	if n.Predecessor == "" || ExclusiveBetween(Hash(address), Hash(n.Predecessor), n.Id) {
		if n.debug {
			fmt.Printf("Notify Success: %s\n", address)
		}
		n.Predecessor = address
		*response = true
	}
	*response = false
	return nil
}

//check whther predecessor is failed
func (n *Node) checkPredecessor() error {
	if err := RPCHealthCheck(n.Predecessor); err != nil {
		n.Predecessor = ""
	}

	return nil
}

//call periodly, refresh.
//n.next is the index of finger to fix
func (n *Node) fixFingerTable() error {
	n.next += 1
	if n.debug {
		fmt.Printf("n.next = %d\n", n.next)
	}
	if n.next >= m {
		n.next = 0
	}

	var response string
	id := fingerEntry(n.Id, n.next)
	if err := n.FindSuccessor(id, &response); err != nil || response == "" {
		fmt.Println("fixFingertable err")
		return err
	}

	for InclusiveBetween(id, n.Id, Hash(response)) {
		n.FingerTable[n.next] = response
		if n.debug {
			fmt.Printf("fixed [%d] = %s\n", n.next, response)
		}
		n.next += 1
		if n.next >= m { ///??????????????????????????????
			n.next = 0
			break
		}
		id = fingerEntry(n.Id, n.next)
	}

	return nil
}

//Call periodly, verfiy succcessor, tell the successor of n
func (n *Node) stabalize() error {
	n.Successor[0] = Addr(n)
	for i := 1; i <= suSize; i++ {
		p, err := RPCGetPredecessor(n.Successor[i])
		if err != nil {
			log.Printf("GetPre %v", err)
			return err
		}
		// a-->c(in a), c-->b(in c), ==> a<-->b<-->c
		// ?? belong (n, successor) ??
		if p != n.Successor[i-1] {
			p, n.Successor[i] = n.Successor[i], p
			//p is now c, s[i] is b, s[i - 1] is a
			if err = RPCNotify(p, n.Successor[i-1]); err != nil {
				return err
			}
		}
	}
}

//using 1 func-- 1 tick strategy, can not sync with(using frequency maybe different)
func (n *Node) stabalizePeriod() error {
	ticker := time.Tick(refreshTime)
	for {
		select { // if <-ticker == time.(0)
		case <-ticker:
			if err := n.stabalize(); err != nil {
				// return err
				fmt.Printf("stabalize err %v\n", err)
			}
		}
	}

	return nil
}
func (n *Node) checkPredecessorPeriod() error {
	ticker := time.Tick(refreshTime)
	for {
		select { // if <-ticker == time.(0)
		case <-ticker:
			if err := n.checkPredecessor(); err != nil {
				fmt.Printf("checkPredecessor err %v\n", err)
				// return err
			}
		}
	}
	return nil
}
func (n *Node) fixFingerTablePeriod() error {
	ticker := time.Tick(refreshTime)
	for {
		select { // if <-ticker == time.(0)
		case <-ticker:
			if err := n.fixFingerTable(); err != nil {
				fmt.Printf("fixFingerTable err %v\n", err)
				// return err
			}
		}
	}

}

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
