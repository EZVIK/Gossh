package service

import "strings"

type SSHDevice struct {
	host       string
	port       string
	username   string
	password   string
	termType   string
	termHeight int
	termWidth  int
	loginLog   bool
	endingArr  []byte
}

func NewDevice(host, username, pass, termType, port string, TermHeight, TermWidth int, LoginLog bool) SSHDevice {
	d := SSHDevice{}
	d.host = host
	d.username = username
	d.password = pass
	d.termHeight = TermHeight
	d.termWidth = TermWidth
	d.loginLog = LoginLog
	d.termType = termType
	d.port = port
	return d
}

func (s *SSHDevice) GetHost() string {
	return s.host
}

func (s *SSHDevice) GetPort() string {
	return s.port
}

// check if output end
func (s *SSHDevice) checkIfEnd(str string) bool {
	str = strings.TrimRight(str, " ")
	for _, end := range s.endingArr {
		if strings.Index(str, string(end)) != -1 {
			return true
		}
	}
	return false
}
