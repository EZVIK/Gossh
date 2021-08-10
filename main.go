package main

import (
	"fmt"
	"github.com/EZVIK/Gossh/SSH_SIM"
	"sync"
	"time"
)

func main() {

	r := new(SSH_SIM.Receiver)
	var wg sync.WaitGroup
	wg.Add(1)

	r.Init(&wg)

	node := SSH_SIM.Node{
		ID:       "1e25006ef76aec99b4403193a244c998",
		TaskId:   "259cf88510e0abd27a317a1cc025d1e3",
		DeviceId: "89c23f5bfad2be62491038f9c4007d3d",
		Command:  "pwd",
		Host:     "159.75.82.148",
		Port:     22,
		Result:   []string{},
	}

	node2 := SSH_SIM.Node{
		ID:       "31602dcfcba525de43a6a70ae72f2198",
		TaskId:   "259cf88510e0abd27a317a1cc025d1e3",
		DeviceId: "89c23f5bfad2be62491038f9c4007d3d",
		Command:  "ls -l /home/nicetry",
		Host:     "159.75.82.148",
		Port:     22,
		Result:   []string{},
	}

	fmt.Println("WTF")
	r.AcceptNode(node)

	time.Sleep(time.Second * 5)
	r.AcceptNode(node2)

	wg.Wait()
}
