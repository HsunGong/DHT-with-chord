package dht

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type Server struct {
	node      *Node
	listener  net.Listener
	listening bool
}

// type Nothing struct{}

func NewServer(n *Node) *Server {
	return &Server{
		node: n, // only one element is inited
	}
}

//address is in s.node
func (s *Server) Listen() {
	//actor--->>>>>>>>>>>>>>>>>>
	rpc.Register(s.node) //using s.node as a object to do things by rpc
	rpc.HandleHttp()

	ler, err := net.Listen("tcp", ":"+s.node.Port) // address
	if err != nil {
		log.Printf("listen error: %v", err)
		panic(err)
	}

	s.node.create()
	s.listener = ler
	s.listening = true
	go http.Serve(l, nil) // goroutine
}

//????server and client 从属关系？？
// avoid repete listening???
func (s *Server) Join(address string) {
	s.Listen()
	s.node.join(address)
}

func (s *Server) Quit() error {
	if err := s.listener.Close(); err != nil {
		s.listening = false
	}
}

func (s *Server) Listening() bool {
	return s.listening
}

func (s *Server) Debug() string {
	return fmt.Sprintf(`
ID: %v
Listening: %v
Address: %v
Data: %v
Successor: %v
Predecessor: %v
Fingers: %v
`, s.node.Id, s.Listening(), s.node.addr(), s.node.Data, s.node.Successor, s.node.Predecessor, s.node.fingers[1:])
}
