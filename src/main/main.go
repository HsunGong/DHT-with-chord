package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"node"
)

const MaxPara = 3

func getline() ([]string, error) {
	//reader := bufio.NewReader(os.Stdin)
	//返回结果包含'\n'？？
	var buffer string
	scanner := bufio.NewScanner(os.Stdin)

	if scanner.Err() != nil {
		fmt.Println("1")
		return []string{}, scanner.Err()
	}

	//_, buffer, err := s
	//scanner.Split()
	if scanner.Scan() {
		buffer = scanner.Text()
		fmt.Println("Order buffers ", buffer)
	}
	return strings.Split(strings.TrimSpace(buffer), " "), nil //delete all ' ' in buffer
}

func main() {
	for {
		log.Printf("dht> ")
		line, err := getline()
		if err != nil {
			fmt.Println("Command format error, get help from command help")
		}

		_, ok := cmd[line[0]]
		if !ok {
			fmt.Println("No such command, get help from command help")
			continue
		}

		err = cmd[line[0]](line[1:]...)

		if err != nil {
			fmt.Println(err)
		}
	}
}

func Port(args ...string) error    { return nil }
func Create(args ...string) error  { return nil }
func Del(args ...string) error     { return nil }
func Ping(args ...string) error    { return nil }
func Put(args ...string) error     { return nil }
func Get(args ...string) error     { return nil }
func Dump(args ...string) error    { return nil }
func Dumpall(args ...string) error { return nil }
func Join(args ...string) error    { return nil }

//func(args...string) error is the cmd funcs return by error
//can't define as const
var cmd = map[string]func(args ...string) error{
	"quit":    Quit,
	"help":    Help,
	"port":    Port,
	"create":  Create,
	"delete":  Del,
	"ping":    Ping,
	"put":     Put,
	"get":     Get,
	"dump":    Dump,
	"dumpall": Dumpall,
	"join":    Join,
}

func Quit(args ...string) error {
	os.Exit(0)
	return nil
}
func Help(args ...string) error {
	var err error
	if len(args) > 1 {
		err = errors.New("Too many arguments. Get help from command help")
	} else {
		err = nil
	}

	switch len(args) {
	case 0:

		fmt.Println(`Commands are:

Current command
	help		displays recognized commands<current command>

Commands related to DHT rings:
	port /<n>	set the listen-on port<n>. (default  3410)
	create		create a new ring.
	join <add>	join an existing ring.
	quit		shut down. This quits and ends the program. 

Commands related to finding and inserting keys and values
	put <k> <v>		insert the given key and value.
	putrandom <n>	randomly generate n <key, value> to insert.
	get <k>			find the given key in the currently active ring. 
	delete <k> 		the peer deletes it from the ring.

Commands that are useful mainly for debugging:
	dump			display information about the current node.
	dumpkey <k>		similar to dump, but this one finds the node resposible for <key>.
	dumpaddr <add>	similar to above, but query a specific host and dump its info.
	dumpall			walk around the ring, dumping all in clockwise order.

Get more details of each command, you can use order <help+command>
eg: help dump, then you will get details of 'dump'
`)
	case 1:

		switch args[0] {
		case "help":
			fmt.Println("the simplest command. This displays a list of recognized commands. Also, the current command")
		case "port":
			fmt.Println(`
port <n> or port
set the port that this node should listen on. 
By default, this should be port 3410, but users can set it to something else.
This command only works before a ring has been created or joined. After that point, trying to issue this command is an error.
`)
		case "create":
			fmt.Println(`
create
create a new ring.
This command only works before a ring has been created or joined. 
After that point, trying to issue this command is an error.
`)
		case "join":
			fmt.Println(`
join <address>
join an existing ring, one of whose nodes is at the address specified.
This command only works before a ring has been created or joined.
After that point, trying to issue this command is an error.
`)
		case "quit":
			fmt.Println(`
quit
shut down.This quits and ends the program. 
If this was the last instance in a ring, the ring is effectively shut down.
If this is not the last instance, it should send all of its data to its immediate successor before quitting. Other than that, it is not necessary to notify the rest of the ring when a node shuts down.
`)
		case "put":
			fmt.Println(`
there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

put <key> <value> 
insert the given key and value into the currently active ring. 
The instance must find the peer that is responsible for the given key using a DHT lookup operation, 
then contact that host directly and send it the key and value to be stored.
`)
		case "putrandom":
			fmt.Println(`
Next, there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

putrandom <n>
randomly generate n keys (and accompanying values) and put each pair into the ring. Useful for debugging.
`)
		case "get":
			fmt.Println(`
Next, there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

get <key>
find the given key in the currently active ring.
The instance must find the peer that is responsible for the given key using a DHT lookup operation, 
then contact that host directly and retrieve the value and display it to the local user.
`)
		case "delete":
			fmt.Println(`
Next, there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

delete <key>
similar to lookup, but instead of retrieving the value and displaying it, the peer deletes it from the ring.

`)
		case "dump":
			fmt.Println(`
For debugging

dump
display information about the current node, including the range of keys it is resposible for,
 its predecessor and successor links, its finger table, and the actual key/value pairs that it stores.
`)
		case "dumpkey":
			fmt.Println(`
For debugging

dumpkey <key>
similar to dump, but this one finds the node resposible for <key>, 
asks it for its dump info, and displays it to the local user. 
This allows a user at one terminal to query any part of the ring.
`)
		case "dumpaddr":
			fmt.Println(`
For debugging

dumpaddr <address>
similar to above, but query a specific host and dump its info.
`)
		case "dumpall":
			fmt.Println(`
For debugging

dumpall
walk around the ring, dumping all information about every peer in the ring in clockwise order 
(display the current host, then its successor, etc).
`)
		default:
			fmt.Println("Wrong command, get help from command help")
		}
	}

	return err
}
