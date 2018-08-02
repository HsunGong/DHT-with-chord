package dht

import (
	// "dht"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/rpc"
)

const (
	DefaultHost = "localHost"
	DefaultPort = "3410"
)

//actually a struct is here:
// client, method, request, response

func Dial(address string) (*rpc.Client, error) {
	if address == "" {
		address = DefaultHost + ":" + DefaultPort
	}
	return rpc.DialHTTP("tcp", address)
}

func Call(address string, method string, request interface{}, response interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Printf("dial failed: %v", err)
		return err
	}
	defer client.Close()

	//get call
	err = client.Call(method, request, response)
	if err != nil {
		log.Printf("call error: %v", err)
		return err
	}

	return nil
}

func RPCNotify(address string, notice_node string) error {
	if address == "" {
		return errors.New("Notify: rpc address is empty")
	}
	response := false

	return Call(address, "Node.Notify", notice_node, &response)
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

func RPCHealthCheck(addr string) error {
	if addr == "" {
		return errors.New("HealthCheck: rpc address is empty")
	}

	response := 0
	if err := Call(addr, "Node.Ping", 101, &response); err != nil {
		return err
	}

	return nil
}

//actually RPCTest
func Testcli(address string, testmsg string) error {
	var size int
	if err := Call(address, "Node.Test", testmsg, &size); err != nil {
		log.Printf("Test rpc Error: %v", err)
		return err
	}
	fmt.Printf("get msg num: %d\n", size)
	return nil
}

func RPCPing(address string) error {
	var response int
	if err := Call(address, "Node.Ping", 3, &response); err != nil {
		return err
	}
	
	fmt.Printf("Got response %d from Ping(3)\n", response)
	return nil
}
