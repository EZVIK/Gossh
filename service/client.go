package service

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"time"
)

type SSHClient struct {
	ID        string
	client    *ssh.Client
	terminal  *Terminal
	device    *SSHDevice
	closeTime time.Time
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

	// new Terminal
	s.terminal = &Terminal{
		session: session,
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
		//Timeout:         time.Second * 300,
	}

	// connect & save client
	s.client, err = ssh.Dial("tcp", s.device.GetHost()+":"+s.device.GetPort(), sshConfig)
	if err != nil {
		return err
	}

	return s.New()
}

func (s *SSHClient) SetTimeOut(duration time.Duration) error {
	s.closeTime = time.Now().Add(duration)
	fmt.Println(s.closeTime.Format("T2006-01-02 15:04:05"))
	return nil
}
