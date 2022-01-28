package core

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strings"
	"time"
)

var defaultTimeout = time.Duration(10) * time.Second

type Gossh struct {
	host       string
	port       int
	user       string
	password   string
	config     *ssh.ClientConfig
	client     *ssh.Client
	session    *ssh.Session
	cliTimeout time.Duration // cli wait timeout
	lifetime   time.Duration // terminal lifetime

	// Shell std
	stdin  io.WriteCloser
	stdout io.Reader
	stderr io.Reader

	Options
}

type CliCommands struct {
	Command []string
	Timeout time.Duration
}

type Options struct {

	// if enable cli output will with user prefix
	// like [root@VM-12-14-centos ~]#
	WithUserPrefix bool

	// process will cut the output acording to this variable
	// like normal user [root@VM-12-14-centos ~]# ,'#'
	// like root   user [root@VM-12-14-centos ~]# ,'#'
	// When the result ends with '#', the result is returned
	OutputCompletedBytes map[string]string
}

func NewSSHClient(host string, port int, cfg *ssh.ClientConfig) *Gossh {

	return &Gossh{
		host:       host,
		port:       port,
		config:     cfg,
		cliTimeout: time.Duration(50000) * time.Millisecond,
	}
}

// Connect to the remote server
func (g *Gossh) Connect() (string, error) {

	var err error

	g.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", g.host, g.port), g.config)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to dial: " + err.Error()))
	}

	g.session, err = g.client.NewSession()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to create session: " + err.Error()))
	}

	if err := redirectStd(g); err != nil {
		return "", err
	}

	loginInfo := read(g, defaultTimeout)

	return loginInfo, nil
}

func (g *Gossh) Close() error {

	if g.session != nil {
		g.session.Close()
	}
	if g.client != nil {
		g.client.Close()
	}

	return nil
}

func (g *Gossh) Exec(commands CliCommands) ([]string, error) {

	var results = make([]string, len(commands.Command))

	results, err := g.run(commands)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (g *Gossh) run(commands CliCommands) ([]string, error) {

	results := make([]string, 0)
	for _, cmd := range commands.Command {

		_, err := g.stdin.Write([]byte(cmd + "\n"))
		if err != nil {
			return nil, err
		}

		if result := read(g, commands.Timeout); err == nil {

			ans := strings.Split(result, "\r\n")
			results = append(results, ans...)
		} else {
			return nil, err
		}
	}

	return results, nil
}

func read(g *Gossh, timout time.Duration) string {
	tk := time.NewTicker(timout)
	ans := ""

	for {
		select {
		// Timeout timer, break loop when timer is <-
		case <-tk.C:
			return ans
		// continue read from std
		default:
			read, err := readOnce(g)
			if err != nil {
				return ""
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

func readOnce(g *Gossh) (string, error) {
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
func redirectStd(g *Gossh) (err error) {
	fd := 0
	_, err = terminal.MakeRaw(fd)
	if err != nil {
		return err
	}

	// termWidth, termHeight affect to the output strings length, width
	// termType affect to the style of the
	termWidth, termHeight, termType := 5000, 50, "vt220"

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
		User: "",
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

	res, err := ssh.Exec(CliCommands{
		Command: []string{
			"docker ps",
		},
		Timeout: time.Duration(5000) * time.Millisecond,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	fmt.Println(res)

}
