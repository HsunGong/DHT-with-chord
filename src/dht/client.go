package dht

import (
	// "dht"
	// "dht"
	"errors"
	"fmt"
	"math/big"
	"net/rpc"
)

const (
	DefaultHost = "localHost"
	DefaultPort = "3410"
)

//actually a struct is here:
// client, method, request, response

func Dial(address string) (*rpc.Cerror error) {
	if address == "" {
		address = DefaultHost + ":" + DefaultPort
	}
	return rpc.DialHTTP("tcp", address)
}

func Call(address string, method string, request interface{}, response interface{}) error {
	if address == ""{
		return errors.New("Call Err: No address access")
	}
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		Logger.Printf("dial failed: %v", err)
		return err
	}
	defer client.Close()

	//get call
	err = client.Call(method, request, response)
	if err != nil {
		Logger.Printf("call error: %v", err)
		return err
	}

	return nil
}

// address.Predecessor --- be_noticed_node -- address
func RPCNotify(address string, new_predecessor string) error {
	if address == "" {
		return errors.New("Notify: rpc address is empty")
	}
	response := false
	// fmt.Println("RPC NOtify get")
	return Call(address, "Node.Notify", new_predecessor, &response)
}

func RPCGetPredecessor(addr string) (string, error) {
	if addr == "" {
		return "", errors.New("GetPredecessor: rpc address is empty")
	}

	response := ""
	if err := Call(addr, "Node.GetPredecessor", false, &response); err != nil {
		return "", err
	}
	if response == "" {
		return "", errors.New("GetPredecessor: rpc Empty predecessor")
	}

	return response, nil
}
func RPCFindSuccessor(addr string, id *big.Int) (string, error) {
	if addr == "" {
		return "", errors.New("RPCFindSuccessor: rpc address is empty")
	}

	response := ""
	if err := Call(addr, "Node.FindSuccessor", id, &response); err != nil {
		return response, err
	}

	return response, nil
}

func RPCPing(address string) error {
	var response int
	if err := Call(address, "Node.Ping", 3, &response); err != nil {
		return err
	}

	fmt.Printf("Got response %d from Ping(3)\n", response)
	return nil
}

func RPCAdapt(n *Node) error {
	err := errors.New("Adapt Init")
	i := 1
	var response bool
	for err != nil && i <= suSize {
		// fmt.Println("aefnasdkjace a----------------------------------------------------------------------------------------->>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
		err = Call(n.SuccessorTable[i], "Node.Adapt", n.Data, response)
		i++
	}

	if err != nil || response != true {
		return err
	}
	return nil
}

//if return "", cant do the function anymore, return errors till main()
//if dial is panic, panic the whole program????????????
func find(address string, key string) string {
	response, err := RPCFindSuccessor(address, Hash(key))
	if err != nil {
		fmt.Printf("find address: %v\n", err) //maybe panic??
		return ""
	}

	return response
}

func RPCPut(address string, key string, val string) error {
	put_node := find(address, key) // key's successor
	if put_node == "" {
		return errors.New("can't get address")
	}

	var response bool
	if err := Call(put_node, "Node.Put", KVP{K: key, V: val}, &response); err != nil {
		return err
	}

	fmt.Printf("Put [%v] stored %v at %s is %t\n", put_node, val, key, response)
	return nil
}
func RPCGet(address string, key string) error {
	get_node := find(address, key) // key's successor
	if get_node == "" {
		return errors.New("can't get address")
	}

	var response string
	if err := Call(get_node, "Node.Get", key, &response); err != nil {
		return err
	}
	fmt.Printf("Get [%v] stored %v at %s\n", get_node, response, key)
	return nil
}

func RPCDel(address string, key string) error {
	del_node := find(address, key) // key's successor
	if del_node == "" {
		return errors.New("can't get address")
	}

	var response bool
	if err := Call(del_node, "Node.Del", key, &response); err != nil {
		return err
	}

	fmt.Printf("Del [%v] KVPair(%v) is %t\n", del_node, key, response)
	return nil
}
func RPCGetSuccessors(address string, s []string) error {
	var a int
	if err := Call(address, "Node.GetSuccessors", a, s); err != nil {
		return err
	}
	return nil
}
