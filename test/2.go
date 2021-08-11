package main

import (
	"github.com/EZVIK/Gossh/SSH_SIM"
	"log"
)

type sshStatus uint32

const (
	Unknown sshStatus = iota
	Wait
	Connected
	UnLink
	end
)

var hostq = "159.75.82.148"

func main() {

	sw := SSH_SIM.Device{
		ID: "31602dcfcba525de43a6a70ae72f2198",
		//TaskId: "259cf88510e0abd27a317a1cc025d1e3",
		//DeviceId: "89c23f5bfad2be62491038f9c4007d3d",
		//Command: "ls -l /home/nicetry",
		Host:     "159.75.82.148",
		Port:     "22",
		Username: "root",
		Password: "elish828MKB",
	}

	if err := sw.Connect(); err != nil {
		log.Fatal(err)
	}

	defer sw.Client.Close()
	defer sw.Session.Close()
	defer sw.Stdin.Close()

	if err := sw.Session.Shell(); err != nil {
		log.Fatal(err)
	}

	commands := []string{"ls -l", "cd /home", "ls -l"}
	if err := sw.SendConfigSet(commands); err != nil {
		log.Fatal(err)
	}

	sw.Session.Wait()

	sw.PrintOutput()
	//sw.PrintErr()
}
