package service

import (
	"fmt"
)

//type ConnectMap struct {
//	connMap map[string]SSHClient
//}

type RuntimeMap struct {
	connMap map[string]*SSHClient
	tmpData map[string]SSHDevice
}

func NewConnectMap() RuntimeMap {
	r := RuntimeMap{}
	r.connMap = make(map[string]*SSHClient, 0)
	r.tmpData = map[string]SSHDevice{
		"159.75.82.148": NewDevice("159.75.82.148", "22", "root", "elish828MKB", "vt220", 500, 40, true),
		"43.128.63.180": NewDevice("43.128.63.180", "22", "root", "elish000MKB", "vt220", 500, 40, true),
	}
	return r
}

// Get Client from map
func (m RuntimeMap) Get(IP string) (*SSHClient, error) {

	c, ok := m.connMap[IP]
	if !ok && c == nil {
		d := m.tmpData[IP]
		c := new(SSHClient)
		c.SetDevice(d)
		m.connMap[IP] = c
		if err := c.Login(); err != nil {
			m.connMap[IP] = nil
			fmt.Printf("IP:%s, Error:%s", IP, err)
			return nil, err
		}
	}

	return c, nil
}
