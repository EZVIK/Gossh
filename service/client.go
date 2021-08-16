package service

import (
	"golang.org/x/crypto/ssh"
	"log"
	"time"
)

type SSHClient struct {
	ID       string
	client   *ssh.Client
	terminal *Terminal
	device   *SSHDevice
}

func NewClient(device SSHDevice) (*SSHClient, error) {
	s := new(SSHClient)
	s.SetDevice(device)
	return s, nil
}

func (s *SSHClient) SetDevice(device SSHDevice) {
	s.device = &device
}

// New Session
func (s *SSHClient) New() error {
	session, err := s.client.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	s.terminal = &Terminal{
		Session: session,
		device:  s.device,
	}

	return s.terminal.interactiveSession(s.device)
}

func (s *SSHClient) Login() (err error) {

	sshConfig := &ssh.ClientConfig{
		User: s.device.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.device.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 300,
	}

	// connect & save client
	s.client, err = ssh.Dial("tcp", s.device.GetHost()+":"+s.device.GetPort(), sshConfig)
	if err != nil {
		return err
	}

	return s.New()
}
