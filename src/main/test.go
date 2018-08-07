package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var (
	buffer bytes.Buffer
)

func init() {

}

func main() {
	var ports [105]string
	data := map[string]string{}
	// var Cmd []command
	Cmd := make([]command, 50, 50)
	// r := rand.New(rand.NewSource(0))

	for i := 0; i <= 100; i++ {
		ports[i] = strconv.FormatInt(int64(i+8000), 10)
	}
	for i := 0; i <= 1500; i++ {
		data[strconv.FormatInt(int64(i), 10)] = strconv.FormatInt(int64(i), 10)
	}

	cnt := 0
	cur := &Cmd[cnt]
	cur.Port(ports[0])
	cnt++
	cur.Create()

	for i := 0; i < 5; i++ {
		for j := 0; j < 4; j++ {
			cur = &Cmd[cnt]
			cur.Port(ports[cnt])
			cnt++
			if err := cur.Join(Cmd[0].node.Address); err != nil {
				fmt.Println(err)
			}
			fmt.Println("Joining  ", j)
			time.Sleep(time.Second)
		}

		time.Sleep(3 * time.Second)
		for j := 0; j < 2; j++ {
			cur = &Cmd[cnt]
			cnt--
			if err := cur.Quit(); err != nil {
				fmt.Println(err)
			}
			time.Sleep(2 * time.Second)
		}
	}

}

func getline(reader io.Reader) ([]string, error) {
	//reader := bufio.NewReader(os.Stdin)
	//返回结果包含'\n'？？
	buffer := make([]string, 0, 10)
	scanner := bufio.NewScanner(reader)

	if scanner.Err() != nil {
		fmt.Println("1")
		return []string{}, scanner.Err()
	}

	//_, buffer, err := s
	f := func(from string, to *[]string) {
		tmp := strings.Split(from, " ")
		for _, s := range tmp {
			if s != "" {
				*to = append(*to, s)
				// fmt.Println(*to)
			}
		}
	}
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		return bufio.ScanLines(data, atEOF)
	}
	scanner.Split(split)
	if scanner.Scan() {
		f(scanner.Text(), &buffer)
		// fmt.Println(buffer)
		// for i, _ := range buffer {
		// 	fmt.Printf("Order buffers '%s'\n", buffer[i])
		// }
	}
	if len(buffer) == 0 {
		return buffer, errors.New("empty line")
	}
	return buffer, nil //delete all ' ' in buffer
}

// func main() {

// 	t := time.Now()
// 	fmt.Printf("@CopyRight(c) 2018 Xun. All rights reserved\n--At %v DHT begins--\n", t.Round(time.Second).Format(layout))

// 	var Cmd [10]command
// 	cnt := 0
// 	port := 8000
// 	for {
// 		c := &Cmd[cnt]

// 		line, _ := getline(os.Stdin)

// 		if line[0] == "create" {
// 			c.Create(line[1:]...)
// 			cnt++
// 		} else if line[0] == "port" {
// 			c.Port(strconv.FormatInt(int64(port+cnt), 10))
// 		} else if line[0] == "dump" {
// 			Cmd[0].Dump(line[1:]...)
// 			Cmd[1].Dump(line[1:]...)
// 		} else if line[0] == "join" {
// 			c.Join(line[1:]...)
// 			cnt++
// 		}
// 	}
// }
