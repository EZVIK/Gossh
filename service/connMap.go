package service

import (
	"errors"
	"fmt"
)

// RuntimeMap Connection Map
type RuntimeMap struct {
	connMap map[string]*SSHClient
	tmpData map[string]SSHDevice
}

func NewConnectMap() RuntimeMap {
	r := RuntimeMap{}
	r.connMap = make(map[string]*SSHClient, 0)
	b := []byte{'#'}
	r.tmpData = map[string]SSHDevice{
		"159.75.82.148": NewDevice("159.75.82.148", "root", "elish828MKB", "vt220", "22", 500, 40, true, b),
		"43.128.63.180": NewDevice("43.128.63.180", "root", "elish000MKB", "vt220", "22", 500, 40, true, b),
	}
	return r
}

// Get Client from map
func (r RuntimeMap) Get(IP string) (*SSHClient, error) {

	craw, ok := r.connMap[IP]
	if !ok && craw == nil {
		d, ok := r.tmpData[IP]
		if !ok {
			return nil, errors.New(fmt.Sprintf("database without this device [%s]", IP))
		}

		c := new(SSHClient)
		c.SetDevice(d)
		r.connMap[IP] = c
		if err := c.Login(); err != nil {
			r.connMap[IP] = nil
			fmt.Printf("IP:[%s], Error:%s", IP, err)
			return nil, err
		}

		return c, nil
	}

	return craw, nil
}

func (r *RuntimeMap) RunCmd(IP string, commands []string) (map[string][]string, error) {

	// get client
	c, err := r.Get(IP)
	if err != nil {
		fmt.Println("get client failed")
		return nil, err
	}

	// Run commands
	ans, err := c.terminal.Run(commands)
	if err != nil {
		fmt.Println("run commands failed")
		return nil, err
	}

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
