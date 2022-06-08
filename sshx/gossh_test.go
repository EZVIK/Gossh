package sshx

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"testing"
)

func Test_ssh_session(t *testing.T) {
	cfg := ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("Elish828"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	cli, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", "159.75.82.148", 22), &cfg)
	if err != nil {
		panic(err)
	}

	s1, err := cli.NewSession()
	if err != nil {
		t.Error(fmt.Errorf("s1 error: %v", err))
		os.Exit(1)
	}

	s1.Wait()

}
