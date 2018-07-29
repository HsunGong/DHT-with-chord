package dht

import (
	"errors"
	"log"
	"math/big"
	"net/rpc"
)

const (
	defaultHost = "localHost"
	defaultPort = "3410"
)

func call(address string, method string, request interface{}, response interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Printf("dial failed: %v", err)
		return err
	}
	defer client.Close()

	err = client.Call(method, request, response)
	if err != nil {
		log.Printf("call error: %v", err)
		return err
	}

	return nil
}

func RPCNotify(addr string, notice string) error {
	//var junk Nothing
	if addr == "" {
		return errors.New("Notify: rpc address is empty")
	}
	response := false

	return call(addr, "Node.Notify", notice, &response)
}

func RPCGetPredecessor(addr string) (string, error) {
	if addr == "" {
		return "", errors.New("GetPredecessor: rpc address is empty")
	}

	response := ""
	if err := call(addr, "Node.GetPredecessor", false, &response); err != nil {
		return "", err
	}
	if response == "" {
		return "", errors.New("GetPredecessor: rpc Empty predecessor")
	}

	return response, nil
}
func RPCFindSuccessor(addr string, id *big.Int) (string, error) {
	if addr == "" {
		return "", errors.New("FindPredecessor: rpc address is empty")
	}

	response := ""
	if err := call(addr, "Node.GetPredecessor", id, &response); err != nil {
		return response, err
	}

	return response, nil
}
func RPCHealthCheck(addr string) (bool, error) {
	if addr == "" {
		return false, errors.New("HealthCheck: rpc address is empty")
	}

	response := 0
	if err := call(addr, "Node.Ping", 101, &response); err != nil {
		return false, err
	}

	return true, nil
}
