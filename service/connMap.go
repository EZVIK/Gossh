package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// RuntimeMap Connection Map
type RuntimeMap struct {
	connMap map[string]*SSHClient
	length  int
	tmpData map[string]SSHDevice
}

func NewConnectMap() RuntimeMap {
	r := RuntimeMap{}
	r.connMap = make(map[string]*SSHClient, 0)
	b := []byte{'#'}
	r.tmpData = map[string]SSHDevice{
		"159.75.82.148": NewDevice("159.75.82.148", "root", "", "vt220", "22", 500, 40, true, b, 30),
		"43.128.63.180": NewDevice("43.128.63.180", "root", "", "vt220", "22", 500, 40, true, b, 30),
	}
	return r
}

// Get Client from map
func (r RuntimeMap) Get(IP string) (*SSHClient, error) {

	existClient, ok := r.connMap[IP]

	if !ok && existClient == nil {
		d, ok := r.tmpData[IP]
		if !ok {
			return nil, errors.New(fmt.Sprintf("database without this device [%s]", IP))
		}

		newClint := new(SSHClient)
		newClint.SetDevice(d)
		r.connMap[IP] = newClint
		if err := newClint.Login(); err != nil {
			r.connMap[IP] = nil
			fmt.Printf("IP:[%s], Error:%s", IP, err)
			return nil, err
		}

		return newClint, nil
	}

	return existClient, nil
}

func (r *RuntimeMap) RunCmd(IP string, cmdReq CMD) (map[string][]string, error) {

	// get client
	c, err := r.Get(IP)
	if err != nil {
		fmt.Println("get client failed")
		return nil, err
	}

	commands := strings.Split(cmdReq.Command, ";:;")
	// Run commands
	ans, err := c.terminal.Run(commands)
	if err != nil {
		fmt.Println("run commands failed")
		return nil, err
	}

	_ = c.SetTimeOut(time.Second * 30)

	return ans, nil
}

func (r *RuntimeMap) CloseAll() {
	for _, t := range r.connMap {
		err := t.client.Close()
		if err != nil {
			return
		}
		fmt.Printf("[%s] closed.\n", t.device.GetHost())
	}
}

func (r *RuntimeMap) GetConnList() []string {
	ans := make([]string, 0)
	for _, v := range r.connMap {
		ans = append(ans, v.device.host)
	}
	return ans
}

func (r *RuntimeMap) CheckClientTimeout() {

	for {

		now := time.Now()
		for _, c := range r.connMap {
			if c.closeTime.Unix() < now.Unix() {
				fmt.Println(c.closeTime.Unix(), now.Unix())

				_ = c.client.Close()
				delete(r.connMap, c.device.GetHost())

				fmt.Println("deleted", c.device.GetHost())
			}
		}

		fmt.Println("Sleep.")
		time.Sleep(10 * time.Second)
	}

}
