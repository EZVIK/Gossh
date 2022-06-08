package test

import (
	"fmt"
	"github.com/EZVIK/Gossh/sshx"
	"golang.org/x/crypto/ssh"
	"log"
	"syscall"
	"testing"
)

func Test_Gossh(t *testing.T) {
	cfg := ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("Elish8281"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	gossh := sshx.NewSSHClient("159.75.82.148", 22, &cfg)

	loginInfo, err := gossh.Connect()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(loginInfo)

}

func Test_operation(t *testing.T) {

	setuidErr := syscall.Setuid(0)
	if setuidErr != nil {
		log.Fatal(setuidErr)
	}

}
