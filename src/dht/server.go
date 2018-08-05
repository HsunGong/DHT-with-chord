package dht

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

var (
	Logger *log.Logger
	debug  bool
	logf   *os.File
)

func init() {
	debug = false
	var err error
	logf, err := os.OpenFile("./logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		log.Panicln("Crash in Log file init")
	}
	Logger = log.New(logf, ">>> ", log.Ltime)
}

//local address
func GetAddress() string {
	var address string
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Panicf("init: GetAddress of finding network: %v", err)
	}

	// find the first non-loopback interface with an IP address
	for _, elt := range interfaces {

		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addres, err := elt.Addrs()
			if err != nil {
				log.Panicf("init: Get addresses of local address: %v", err)
			}

			for _, addr := range addres {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						address = ip4.String()
						break
					}
				}
			}
		}
	}
	if address == "" {
		log.Panicf("init: failed to find non-loopback interface with valid address on this node")
	}

	return address
}

type Server struct {
	node     *Node
	listener net.Listener
	// listening bool

	// logfile *os.File
}

// type Nothing struct{}
// 通常情况下在程序日志里记录一些比较有意义的状态数据：
// 程序启动，退出的时间点；程序运行消耗时间；
// 耗时程序的执行进度；重要变量的状态变化。
// 初次之外，在公共的日志里规避打印程序的调试或者提示信息。

func NewServer(n *Node) *Server {
	//什么时候消亡？？
	p := &Server{
		node: n, // only one element is inited
		// listening: false,
	}
	Logger.Println("--------------------------------------------------------------- <<<")
	Logger.Println("Init a Server at ", time.Now())
	return p
}

//address is in s.node
//this is the function to run a server
func (s *Server) Listen() error {
	// if s.listening {
	// return errors.New("Already listening")
	// }

	//actor--->>>>>>>>>>>>>>>>>>
	rpc.Register(s.node) //using s.node as a object to do things by rpc
	rpc.HandleHTTP()

	ler, err := net.Listen("tcp", ":"+s.node.Port) // address
	if err != nil {
		fmt.Printf("listen error: %v", err)
		Logger.Panicf("listen error: %v", err)
		//panic(err)
	} else {
		Logger.Println("listen at ", ":"+s.node.Port)
	}

	s.node.create()
	s.listener = ler
	// s.listening = true

	go http.Serve(ler, nil) // goroutine
	return nil
}

// avoid repete listening???
// the port is server.node.port(not port outside)
func (s *Server) Join(address string) error {

	if err := s.Listen(); err != nil {
		return err
	}
	return s.node.join(address)
}

//for a server, it means unlisten
func (s *Server) Quit() error {
	if err := s.listener.Close(); err != nil {
		return err
	}

	if err := RPCAdapt(s.node); err != nil {
		return err
	}

	if err := logf.Close(); err != nil {
		fmt.Printf("logs Close: %v", err)
	}
	return nil
}

func (s *Server) IsListening() bool {
	return s.listener != nil
	// return s.listening
}

func (s *Server) Debug() string {

	return fmt.Sprintf(`
ID: %v
Listening: %v
Address: %v
Data: %v
Successors: %v
Successor: %v
Predecessor: %v
Fingers: %v
`, Hash(s.node.Address), s.IsListening(), s.node.Address, s.node.Data, s.node.SuccessorTable, s.node.successor, s.node.Predecessor, s.node.FingerTable)
}
