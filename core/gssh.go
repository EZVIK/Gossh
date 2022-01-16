package main

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Gssh struct {
	host     string
	port     int
	user     string
	password string
	config   *ssh.ClientConfig
	client   *ssh.Client
	session  *ssh.Session
	timeout  time.Duration

	// Shell std
	stdin  io.WriteCloser
	stdout io.Reader
	stderr io.Reader
}

type CliCommands struct {
	Name    string
	Desc    string
	Command []string
	Result  []string
}

func NewSSHClient(host string, port int, cfg *ssh.ClientConfig) *Gssh {

	return &Gssh{
		host:    host,
		port:    port,
		config:  cfg,
		timeout: time.Duration(50000) * time.Millisecond,
	}
}

func (g *Gssh) Connect() (string, error) {

	var err error
	g.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", g.host, g.port), g.config)
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to dial: " + err.Error()))
	}

	g.session, err = g.client.NewSession()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to create session: " + err.Error()))
	}

	if err := g.redirectStd(); err != nil {
		return "", err
	}

	loginInfo := g.read()

	return loginInfo, nil
}

func (g *Gssh) Exec(commands CliCommands) ([]string, error) {

	var results = make([]string, len(commands.Command))

	for i, cmd := range commands.Command {
		if result, err := g.run(cmd); err == nil {
			results[i] = result
		} else {
			return nil, err
		}
	}

	return results, nil
}

func (g *Gssh) run(command string) (string, error) {

	_, err := g.stdin.Write([]byte(command + "\n"))

	if err != nil {
		return "", err
	}

	result := g.read()

	return result, nil
}

func (g *Gssh) read() string {
	tk := time.NewTicker(g.timeout)
	ans := ""

	for {
		select {
		// Timeout timer, break loop when timer is <-
		case <-tk.C:
			return ans
		// continue read from std
		default:
			read, err := g.readOnce()
			if err != nil {
				log.Fatal(err)
			}
			ans += read

			// check stdout finished output
			if checkIfEnd(read) {
				tmp := ans
				ans = ""
				return tmp
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (g *Gssh) readOnce() (string, error) {
	ans := ""
	buf := make([]byte, 512)
	n, stdoutErr := g.stdout.Read(buf)

	if stdoutErr != nil {
		return ans, stdoutErr
	}

	if n > 0 {
		rawStr := string(buf[:n])
		ans += fmt.Sprintf("%s", rawStr)
		return ans, nil
	}
	return "", nil
}

// return stdin, stdout, stderr
func (g *Gssh) redirectStd() (err error) {
	fd := 0
	_, err = terminal.MakeRaw(fd)
	if err != nil {
		return err
	}

	// termWidth, termHeight affect to the output strings length, width
	// termType affect to the style of the
	termWidth, termHeight, termType := 5000, 50, "vt220"

	//
	err = g.session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	// redirect std
	g.stdin, err = g.session.StdinPipe()
	if err != nil {
		return err
	}

	g.stdout, err = g.session.StdoutPipe()
	if err != nil {
		return err

	}
	g.stderr, err = g.session.StderrPipe()

	err = g.session.Shell()
	if err != nil {
		return err
	}

	return nil
}

func checkIfEnd(str string) bool {
	str = strings.TrimRight(str, " ")
	for _, end := range []byte{'#'} {
		if strings.Index(str, string(end)) != -1 {
			return true
		}
	}
	return false
}

func main() {

	cfg := ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(""),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	ssh := NewSSHClient("", 22, &cfg)

	loginInfo, err := ssh.Connect()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	fmt.Print(loginInfo)
}
