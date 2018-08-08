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
	refreshTime = 100 * time.Millisecond // adjust to avoid error:connectex: Only one usage of each socket address (protocol/network address/port) is normally permitted. in Stabilize with Notify err:
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
		next: -1,
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

func (n *Node) Besplited(address string, response *bool) error {
	// if n.Port == "8003" {
	// 	n.dump()
	// 	fmt.Println("WTF:::: ", address)
	// }

	id := Hash(address)
	ip := Hash(n.Address)
	for key, val := range n.Data {
		if InclusiveBetween(id, Hash(key), ip) {
			if err := Call(address, "Node.Put", KVP{K: key, V: val}, &response); err != nil {
				fmt.Printf("Splited err: %v\n", err)
			}
			delete(n.Data, key)
			// fmt.Printf("split %s to %s: %s, %s\n", n.Address, address, key, val)
		}
	}

	return nil
}

func (n *Node) merge() {
	var response bool
	for key, val := range n.Data {
		for err := Call(n.successor, "Node.Put", KVP{K: key, V: val}, &response); err != nil; err = Call(n.successor, "Node.Put", KVP{K: key, V: val}, &response) {
			// n.dump()
			// fmt.Printf("Merge err : %v\n", err)
		}
		// fmt.Printf("merge %s to %s: %s, %s\n", n.Address, n.successor, key, val)
	}
	return
}

//init node's info
func (n *Node) join(address string) error {
	n.Predecessor = "" // have to be stablized later
	addr, err := RPCFindSuccessor(address, Hash(n.Address))
	if err != nil {
		fmt.Printf("node join-findsuccessor %v\n", err)
		fmt.Println("Address is ", n.Address, address, addr)
		return err
	}
	n.successor = addr
	if err := n.fixSuccessorTable(); err != nil {
		return err
	}

	//notify the successor[0]
	if err := RPCNotify(n.successor, n.Address); err != nil {
		fmt.Println("Join but Notify err")
	}
	// if n.Port == "8006" { //} || n.Port == "8003" {
	// 	n.dump()
	// }

	var response bool
	if err := Call(n.successor, "Node.Besplited", n.Address, &response); err != nil {
		fmt.Println("split err")
	}
	// if n.Port == "8006" {
	// 	n.dump()
	// }

	// n.stabilize()

	return nil
}

func (n *Node) CopySuccessor(a bool, table *([suSize + 1]string)) error {
	// fmt.Println("From ", n.Address, n.SuccessorTable)
	for i := 0; i <= suSize; i++ {
		// fmt.Printf("change %s to %s\n", (*table)[i], n.SuccessorTable[i])
		table[i] = n.SuccessorTable[i]
	}
	// fmt.Println("Copy2 ", table)
	return nil
}

func (n *Node) fixSuccessorTable() error {
	var a bool
	// fmt.Println(">>>Before ", n.successor, n.SuccessorTable)
	if err := Call(n.successor, "Node.CopySuccessor", a, &n.SuccessorTable); err != nil {
		return err
	}
	// fmt.Println("Mid ", n.successor, n.SuccessorTable)

	for i := suSize - 1; i >= 0; i-- {
		n.SuccessorTable[i+1] = n.SuccessorTable[i]
	}
	n.SuccessorTable[0] = n.successor
	// fmt.Println("Copy ", n.successor, n.SuccessorTable)
	return nil
}

// search the local table for the highest predecessor of id
// n.closest_preceding_node(id)
//if empty, return "", left to let
// find finger[i] ∈ (n, id)
func (n *Node) findFingerTable(id *big.Int) string {
	for i := m - 1; i >= 0; i-- {
		if _, err := RPCPing(n.FingerTable[i]); err != nil {
			continue
		}

		//倒着往前找，找到第一个finger使得finger在id和address之间时，此finger的下一个就是我们要找的pre——node
		//或者说，正着找，找到第一个finger使得finger是id的一个假后继，然后前一个finger就是pre——node
		if ExclusiveBetween(Hash(n.FingerTable[i]), Hash(n.Address), id) {
			// the return should be i - 1
			// what if i == 0?, impossible for successor have to be included

			return n.FingerTable[i]
		}
	}
	// return n.FingerTable[m - 1]
	// n.dump()
	log.Printf("wrong")
	//i==0 error

	// all of them is included by n.Id and id, so as said should return n
	//however incase loop forever, return "", let the function to handle
	return n.Address ///??????

}

//If the id I'm looking for falls between me and mySuccessor
//Then the data for this id will be found on mySuccessor
// ------------>>>will do finger table later
// ask node n to find the successor of id
// or a better node to continue the search with
//!!!!!maybe can repete with some node(in fixfinger)
func (n *Node) FindSuccessor(id *big.Int, successor *string) error {
	// fmt.Println(Hash(n.Address), id, n.Address)
	if _, err := RPCPing(n.successor); err == nil && InclusiveBetween(id, Hash(n.Address), Hash(n.successor)) {
		*successor = n.successor
		// fmt.Println(id, " successor find:", n.Address, n.successor)
		return nil
	}

	for i := 1; i <= suSize; i++ {
		// fmt.Println("circle")
		if _, err := RPCPing(n.SuccessorTable[i]); err == nil && InclusiveBetween(id, Hash(n.SuccessorTable[i-1]), Hash(n.SuccessorTable[i])) { //id ∈ (n, successor]
			*successor = n.SuccessorTable[i]
			// fmt.Println(id, " successor find: ", n.SuccessorTable[i-1], n.SuccessorTable[i])
			return nil
		}
	}

	//wont loop if successor is exists
	// fmt.Println(n)

	nextAddr := n.findFingerTable(id)
	// fmt.Printf("[%s]Finger :%s for %d\n", n.Address, nextAddr, id)
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
		// if n.Port == "10.163.174.211:8004" {
		// 	fmt.Println("8004 port is notifying")
		// 	n.dump()
		// 	fmt.Printf("%s Notify %s\n", n.Address, new_predecessor)
		// }
		n.Predecessor = new_predecessor
		*response = true
		return nil
	}
	*response = false
	return nil
}

//check whther predecessor is failed
func (n *Node) checkPredecessor() {
	if _, err := RPCPing(n.Predecessor); err != nil {
		n.Predecessor = ""
	}
}

//call periodly, refresh.
//n.next is the index of finger to fix
func (n *Node) fixFingerTable() {
	var response string
	for true {
		n.next++
		if n.next >= m {
			n.next = -1
			return
		}
		id := fingerEntry(Hash(n.Address), n.next)

		if response == "" {
			if err := n.FindSuccessor(id, &response); err != nil || response == "" {
				fmt.Printf("fixFingertable err at: %v\n", err)
				// fmt.Println(n.next, n.Address, response)
				return
			}
			// color.Yellow("Successor")
		}

		if InclusiveBetween(id, Hash(n.Address), Hash(response)) {
			n.FingerTable[n.next] = response
		} else {
			n.next--
			return
		}
		// os.Exit(1)
	}
}

//Call periodly, verfiy succcessor, tell the successor of n
//only check successor 1
func (n *Node) stabilize() {
	for _, cur := range n.SuccessorTable {
		if _, err := RPCPing(cur); err != nil {
			continue
		}

		if pre_i, err := RPCGetPredecessor(cur); err == nil {
			if ExclusiveBetween(Hash(pre_i), Hash(n.Address), Hash(cur)) {
				n.successor = pre_i
			}
		} else {
			// fmt.Printf("%s stabilize[%d] err %v\n", n.Address, i, err)
			// n.dump()
		}
		if err := n.fixSuccessorTable(); err != nil {
			n.successor = n.SuccessorTable[0]
			// RPCNotify(n.successor, Addr(n))
			fmt.Printf("[%s]Stabilize fix SuccessoTable %v\n", n.Address, err)
			continue
		}

		if err := RPCNotify(n.successor, n.Address); err != nil {
			fmt.Printf("[%s]Stabilize Notify %v\n", n.Address, err)
		}

		return
	}
	// n.dump()
	log.Println("Successor List Failed")

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
	// fmt.Println("Put ", n.Address, args.K, args.V)
	n.Data[args.K] = args.V
	*success = true
	return nil
}
func (n *Node) Get(key string, response *string) error {
	// fmt.Println("Get ", n.Address, key, n.Data[key])
	*response = n.Data[key]
	return nil
}
func (n *Node) Del(key string, success *bool) error {
	if _, ok := n.Data[key]; ok {
		// fmt.Println("Del ", n.Address, key, n.Data[key])
		delete(n.Data, key)
		*success = true
		return nil
	} else {
		*success = false
		return nil //ok func, but can't find key
	}
}

func (n *Node) Ping(request int, response *int) error {
	//do something with request?? no need
	*response = len(n.Data)
	return nil
}

func (node *Node) dump() {
	fmt.Printf(`
ID: %v		Address: %v
S: %v		Successors: %v
Next: %v	Predecessor: %v
Data: %v
Finger: %v
`, Hash(node.Address), node.Address, node.successor, node.SuccessorTable, node.next, node.Predecessor, node.Data, node.FingerTable[:60])
}
