package main

import (
	"bufio"
	"fmt"
	"github.com/EZVIK/Gossh/SSH_SIM"
	"io"
	"log"
)

var (
	username = "root"
	password = "elish828MKB"
	ip       = "159.75.82.148"
	port     = 22
	key      = ""
	cmd      = "ls -l;cd /home;ls -l"
)

func main() {
	//fmt.Println(11076.12 - 10837.97)

	device, node := GetTestData()

	err := device.Connect()
	if err != nil {
		log.Fatal(err)
	}

	session, err := device.Client.NewSession()

	if err != nil {
		log.Fatal(err)
	}

	if err = session.Shell(); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = session.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	_, err = io.WriteString(device.Stdin, node.Command+"\n")
	if err != nil {
		log.Fatal(err)
	}

	session.Wait()

	func() {
		r := bufio.NewReader(device.Stdout)
		for {
			text, err := r.ReadString('\n')
			fmt.Printf("%s", text)
			if err == io.EOF {
				break
			}
		}
	}()
}

func GetTestData() (SSH_SIM.Device, SSH_SIM.Node) {
	return SSH_SIM.Device{
			ID: "31602dcfcba525de43a6a70ae72f2198",
			//TaskId: "259cf88510e0abd27a317a1cc025d1e3",
			//DeviceId: "89c23f5bfad2be62491038f9c4007d3d",
			//Command: "ls -l /home/nicetry",
			Host:     "159.75.82.148",
			Port:     "22",
			Username: username,
			Password: password,
		}, SSH_SIM.Node{
			ID:       "31602dcfcba525de43a6a70ae72f2198",
			TaskId:   "259cf88510e0abd27a317a1cc025d1e3",
			DeviceId: "89c23f5bfad2be62491038f9c4007d3d",
			Command:  "ls -l /home/nicetry",
			Host:     "159.75.82.148",
			Port:     22,
			Result:   []string{},
		}
}

//
//func connect(user, password, host, key string, port int, cipherList []string) (*ssh.Session, error) {
//	var (
//		auth         []ssh.AuthMethod
//		addr         string
//		clientConfig *ssh.ClientConfig
//		client       *ssh.Client
//		config       ssh.Config
//		session      *ssh.Session
//		err          error
//	)
//	// get auth method
//	auth = make([]ssh.AuthMethod, 0)
//	if key == "" {
//		auth = append(auth, ssh.Password(password))
//	} else {
//		pemBytes, err := ioutil.ReadFile(key)
//		if err != nil {
//			return nil, err
//		}
//
//		var signer ssh.Signer
//		if password == "" {
//			signer, err = ssh.ParsePrivateKey(pemBytes)
//		} else {
//			signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
//		}
//		if err != nil {
//			return nil, err
//		}
//		auth = append(auth, ssh.PublicKeys(signer))
//	}
//
//	if len(cipherList) == 0 {
//		config = ssh.Config{
//			Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
//		}
//	} else {
//		config = ssh.Config{
//			Ciphers: cipherList,
//		}
//	}
//
//	clientConfig = &ssh.ClientConfig{
//		User:    user,
//		Auth:    auth,
//		Timeout: 30 * time.Second,
//		Config:  config,
//		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
//			return nil
//		},
//	}
//
//	// connet to ssh
//	addr = fmt.Sprintf("%s:%d", host, port)
//
//	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
//		return nil, err
//	}
//
//	// create session
//	if session, err = client.NewSession(); err != nil {
//		return nil, err
//	}
//
//	modes := ssh.TerminalModes{
//		ssh.ECHO:          0,     // disable echoing
//		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
//		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
//	}
//
//	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
//		return nil, err
//	}
//
//	return session, nil
//}
//
//func Test_SSH(t *testing.T) {
//	var cipherList []string
//	session, err := connect(username, password, ip, key, port, cipherList)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	defer session.Close()
//
//	cmdlist := strings.Split(cmd, ";")
//	stdinBuf, err := session.StdinPipe()
//	if err != nil {
//		t.Error(err)
//		return
//	}
//
//	var outbt, errbt bytes.Buffer
//	session.Stdout = &outbt
//
//	session.Stderr = &errbt
//	err = session.Shell()
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	for _, c := range cmdlist {
//		c = c + "\n"
//		stdinBuf.Write([]byte(c))
//	}
//	session.Wait()
//	t.Log(outbt.String() + errbt.String())
//	fmt.Println(outbt.String() + errbt.String())
//	return
//}
//
