package SSH_SIM

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
)

type Device struct {
	ID       string
	Host     string
	Port     string
	Username string
	Password string
	Key      string
	Client   *ssh.Client
	Session  *ssh.Session
	Stdin    io.WriteCloser
	Stdout   io.Reader
	Stderr   io.Reader
	CurrNode Node
}

func (d *Device) Connect() error {

	client, err := ssh.Dial("tcp", d.Host+":"+d.Port, &ssh.ClientConfig{
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc", "3des-cbc", "blowfish-cbc", "aes256-ctr"},
		},
		User:            d.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(d.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	sshIn, err := session.StdinPipe()
	if err != nil {
		return err
	}
	sshOut, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	sshErr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	d.Client = client
	d.Session = session

	d.Stdin = sshIn
	d.Stdout = sshOut
	d.Stderr = sshErr

	go d.PrintOutput()
	go d.PrintErr()

	return nil
}

func (d *Device) SendCommand(cmd string) (string, error) {
	if _, err := io.WriteString(d.Stdin, cmd+"\n"); err != nil {
		return "", err
	}

	return "", nil
}

func (d *Device) SendConfigSet(cmds []string) error {

	for _, cmd := range cmds {
		if _, err := io.WriteString(d.Stdin, cmd+"\n"); err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) PrintOutput() {
	r := bufio.NewReader(d.Stdout)
	for {
		text, err := r.ReadString('\n')
		fmt.Printf("%s", text)
		if err == io.EOF {
			break
		}
	}
}

func (d *Device) PrintErr() {
	r := bufio.NewReader(d.Stderr)
	for {
		text, err := r.ReadString('\n')
		fmt.Printf("ERROR %s\n", text)
		if err == io.EOF {
			break
		}
	}
}

//// Device device info
//type Device struct {
//	ID          string
//	Client  	*ssh.Client
//	Sessions 	map[string]*ssh.Session			// map[taskId]
//	Username	string
//	Password    string
//	IP      	string
//	PORT		string
//}
//
//// GetDevice fake
//func GetDevice(node Node) (*Device, *ssh.Session) {
//
//	d := new(Device)
//	d.Sessions = make(map[string]*ssh.Session, 0)
//
//	// GET FROM DATABASE
//	clientConfig := ssh.ClientConfig{
//		User: "root",
//		Auth: []ssh.AuthMethod{
//			ssh.Password("elish828MKB"),
//		},
//		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
//	}
//	connection, err := ssh.Dial("tcp", node.GetHostAndPort(), &clientConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	d.Client = connection
//
//	// defer d.Client.Close()
//
//	s, err := d.Client.NewSession()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	d.Sessions[node.TaskId] = s
//
//	return d, s
//}
