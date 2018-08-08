package main

import (
	"bufio"
	"bytes"
	"dht"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/color"
)

var (
	buffer bytes.Buffer
)

func init() {

}

type errType struct {
	No   int    // order in the loop
	node string // join node
	k    string
	v    string
	// cnt  int //data-cnt
}

type sl []*big.Int

// func main() {
// 	x := make(sl, 30, 30)
// 	j := 0
// 	for i := 0; i <= 11; i++ {
// 		x[j] = dht.Hash("10.163.174.211:800" + strconv.FormatInt(int64(i), 10))
// 		j++
// 		fmt.Printf("%d\t%d\n", i, x[j])
// 		x[j] = dht.Hash(strconv.FormatInt(int64(i), 10))
// 		fmt.Printf("%d:\t%d\n", i, x[j])
// 		j++
// 	}
// 	// sort.Sort(x)
// }

func main() {
	var ports [105]string
	datas := map[string]string{}
	var dataString [2005]string
	Cmd := make([]command, 100, 100)

	r := rand.New(rand.NewSource(0))
	// r := rand.New(rand.NewSource(time.Now().Unix()))
	green := color.New(color.FgGreen) //.Add(color.Underline)
	red := color.New(color.FgRed)
	blue := color.New(color.FgBlue)

	for i := 0; i <= 100; i++ {
		ports[i] = strconv.FormatInt(int64(i+6000), 10)
	}

	for i := 0; i <= 2000; i++ {
		dataString[i] = strconv.FormatInt(r.Int63(), 10)
		datas[dataString[i]] = strconv.FormatInt(int64(i), 10)
	}

	cnt := -1   //stack top of porgram
	d_cnt := -1 //stack top of data
	var flag bool
	errs := make([]errType, 0, 300)

	cnt++
	cur := &Cmd[cnt]
	cur.Port(ports[0])
	cur.Create()

	// fmt.Printf("%d\n%d\n%d", dht.Hash("10.163.174.211:8005"), dht.Hash("2"), dht.Hash("10.163.174.211:8010"))
	for i := 1; i <= 5; i++ {
		for j := 1; j <= 15; j++ {
			cnt++
			cur = &Cmd[cnt]
			cur.Port(ports[cnt])
			// if ports[cnt] == "8006" {
			// 	red.Println("here it is")
			// 	for j := 0; j <= cnt; j++ {
			// 		Cmd[j].Dump()
			// 	}

			// }
			if err := cur.Join(Cmd[0].node.Address); err != nil {
				fmt.Println(err)
			}

			blue.Println("Join  ", cnt)
			time.Sleep(1000 * time.Millisecond) // 1000+
		}
		time.Sleep(4000 * time.Millisecond) // 5000+

		//put
		flag = true
		errs = errs[0:0]
		for j := 1; j <= 300; j++ {
			// c := r.Int() % (cnt + 1)
			c := 0
			d_cnt++
			s := dataString[d_cnt]
			Cmd[c].Put(s, datas[s])

			if ans, _ := buffer.ReadString('\n'); ans != "true\n" {
				fmt.Printf("%s is %d\n", s, dht.Hash(s))
				errs = append(errs, errType{j, Cmd[c].node.Address, s, datas[s]})
				flag = false
			}
		}
		if flag {
			green.Println("Pass First ", i)
		} else {
			red.Println("Errors(1) are:")
			for _, j := range errs {
				red.Println(j)
			}
		}

		//get
		flag = true
		errs = errs[0:0]
		for j := 1; j <= 200; j++ {
			// c := r.Int() % (cnt + 1)
			c := 0
			s := dataString[r.Int()%d_cnt]
			Cmd[c].Get(s)
			if ans, _ := buffer.ReadString('\n'); ans != datas[s]+"\n" {
				fmt.Printf("Get(%s) hash:%d\n", s, dht.Hash(s))
				// fmt.Printf("%s is %d\n", s, dht.Hash(s))
				errs = append(errs, errType{j, Cmd[c].node.Address, s, datas[s]})
				flag = false
			}
		}
		if flag {
			green.Println("Pass Second ", i)
		} else {
			red.Println("Errors(2) are:")
			for _, j := range errs {
				red.Println(j)
			}
			// for j := 0; j <= cnt; j++ {
			// 	Cmd[j].Dump()
			// }
			// fmt.Println("\n")
		}

		//del
		flag = true
		errs = errs[0:0]
		for j := 1; j <= 150; j++ {
			// c := r.Int() % (cnt + 1)
			c := 0
			s := dataString[d_cnt]
			d_cnt--
			Cmd[c].Del(s)

			if ans, _ := buffer.ReadString('\n'); ans != "true\n" {
				fmt.Printf("Del(%s) hash:%d\n", s, dht.Hash(s))
				errs = append(errs, errType{j, Cmd[c].node.Address, s, datas[s]})
				flag = false
			}
		}
		if flag {
			green.Println("Pass Third ", i)
		} else {
			red.Println("Errors(3) are:")
			for _, j := range errs {
				fmt.Println("Get hash:", j.k)
				red.Println(j)
			}
			// for j := 0; j <= cnt; j++ {
			// 	Cmd[j].Dump()
			// }
		}

		for j := 1; j <= 5; j++ {
			cur = &Cmd[cnt]
			cnt--
			if err := cur.Quit(); err != nil {
				fmt.Println(err)
			}
			blue.Println("Quit ", cnt+1)
			time.Sleep(1000 * time.Millisecond) // 1000+
		}
		time.Sleep(4000 * time.Millisecond)
		// Cmd[3].Dump()

		//put
		flag = true
		errs = errs[0:0]
		for j := 1; j <= 300; j++ {
			c := r.Int() % (cnt + 1)
			d_cnt++
			s := dataString[d_cnt]
			Cmd[c].Put(s, datas[s])
			datas[s] = strconv.FormatInt(int64(d_cnt), 10)
			if ans, _ := buffer.ReadString('\n'); ans != "true\n" {
				fmt.Printf("%s is %d\n", s, dht.Hash(s))
				errs = append(errs, errType{j, Cmd[c].node.Address, s, datas[s]})
				flag = false
			}
		}
		if flag {
			green.Println("Pass First ", i)
		} else {
			red.Println("Errors(1) are:")
			for _, j := range errs {
				red.Println(j)
			}
		}

		//get
		flag = true
		errs = errs[0:0]
		for j := 1; j <= 200; j++ {
			c := r.Int() % (cnt + 1)
			s := dataString[r.Int()%d_cnt]
			Cmd[c].Get(s)
			if ans, _ := buffer.ReadString('\n'); ans != datas[s]+"\n" {
				fmt.Printf("Get(%s) hash:%d\n", s, dht.Hash(s))
				errs = append(errs, errType{j, Cmd[c].node.Address, s, datas[s]})
				flag = false
			}
		}
		if flag {
			green.Println("Pass Second ", i)
		} else {
			red.Println("Errors(2) are:")
			for _, j := range errs {
				red.Println(j)
			}
		}

		//del
		flag = true
		errs = errs[0:0]
		for j := 1; j <= 150; j++ {
			c := r.Int() % (cnt + 1)
			s := dataString[d_cnt]
			d_cnt--
			Cmd[c].Del(s)

			if ans, _ := buffer.ReadString('\n'); ans != "true\n" {
				fmt.Printf("Del(%s) hash:%d\n", s, dht.Hash(s))
				errs = append(errs, errType{j, Cmd[c].node.Address, s, datas[s]})
				flag = false
			}
		}
		if flag {
			green.Println("Pass Third ", i)
		} else {
			red.Println("Errors(3) are:")
			for _, j := range errs {
				fmt.Println("Get hash:", j.k)
				red.Println(j)
			}
		}
	}
	green.Println(cnt, d_cnt+1)

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
// 		} else if line[0] == "put" {
// 			c.Put(line[1:]...)
// 		} else if line[0] == "get" {
// 			c.Get(line[1:]...)
// 			var s string
// 			fmt.Fscanln(&buffer, s)
// 			fmt.Println(s)
// 		} else if line[0] == "del" {
// 			c.Del(line[1:]...)
// 		}

// 	}
// }
