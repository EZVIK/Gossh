package sshx

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"strings"
	"time"
)

var (
	defaultTimeout = time.Duration(10) * time.Second

	defaultReadGapTime = time.Duration(10) * time.Millisecond
)

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

	Option
}

type CliCommands struct {
	Command []string
	Timeout time.Duration
	//Option
}

//type Option struct {
//
//	// if enable cli output will with user prefix
//	// like [root@VM-12-14-centos ~]#
//	WithUserPrefix bool
//
//	// process will cut the output acording to this variable
//	// like normal user [root@VM-12-14-centos ~]# ,'#'
//	// like root   user [root@VM-12-14-centos ~]# ,'#'
//	// When the result ends with '#', the result is returned
//	OutputCompletedBytes map[string]string
//
//	// time of waiting Cli stdout
//	CliTimeout time.Duration
//
//	// gap time of read cli stdout
//	ReadGapTime time.Duration
//}

// NewSSHClient return Gossh
func NewSSHClient(host string, port int, cfg *ssh.ClientConfig, opts ...Option) *Gossh {

	//var opt options
	//for _, o := range opts {
	//	o(&opt)
	//}

	return &Gossh{
		host:       host,
		port:       port,
		config:     cfg,
		cliTimeout: time.Duration(50000) * time.Millisecond,
	}
}

// Connect to the remote server
// return login text
func (g *Gossh) Connect() (string, error) {

	var err error

	// create connection
	g.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", g.host, g.port), g.config)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to dial: " + err.Error()))
	}

	g.session, err = g.client.NewSession()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to create session: " + err.Error()))
	}

	// redirect std
	if err = redirectStd(g); err != nil {
		return "", err
	}

	loginInfo := read(g, defaultTimeout)

	return loginInfo, nil
}

// Close connection
func (g *Gossh) Close() error {

	if g.session != nil {
		g.session.Close()
	}
	if g.client != nil {
		g.client.Close()
	}

	return nil
}

// Exec CLiCommands
// return multi result
func (g *Gossh) Exec(commands CliCommands) (map[string][]string, error) {

	var results = make(map[string][]string, len(commands.Command))

	results, err := g.run(commands, func(s string) []string {
		return strings.Split(s, "\r\n")
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// run commands and read result from stdout
func (g *Gossh) run(commands CliCommands, textProcessFunc func(string) []string) (map[string][]string, error) {

	results := make(map[string][]string, 0)
	for _, cmd := range commands.Command {

		// input cli
		_, err := g.stdin.Write([]byte(cmd + "\n"))
		if err != nil {
			return nil, err
		}

		// read from stdout
		if result := read(g, commands.Timeout); err == nil {

			results[cmd] = textProcessFunc(result)

		} else {
			return nil, err
		}
	}

	return results, nil
}

func read(g *Gossh, timeout time.Duration) string {

	// read timeout ticker
	tk := time.NewTicker(timeout)
	ans := ""

	for {
		select {
		// Timeout timer, break loop when timer is <-
		case <-tk.C:
			return ans

		// continue read from std
		default:
			stream, err := readOnce(g, func(s string) string {

				// optional function
				res := strings.Replace(s, " ", "", 0)
				if res == "" {
					return ""
				}
				return s

			})
			if err != nil {
				return ""
			}

			ans += stream

			// check stdout finished output
			if checkIfEnd(stream) {
				tmp := ans
				ans = ""
				return tmp
			}
			time.Sleep(defaultReadGapTime)

		}
	}
}

// read from std fd
func readOnce(g *Gossh, filter func(string) string) (string, error) {
	ans := ""
	buf := make([]byte, 1024)
	n, stdoutErr := g.stdout.Read(buf)

	if stdoutErr != nil {
		return ans, stdoutErr
	}

	if n > 0 {
		//fmt.Println("[" + string(buf) + "]")
		//fmt.Println("{" + string(buf[:n]) + "}")
		str := string(buf[:n])
		res := filter(str)
		res = strings.Trim(res, " ")

		return res, nil
	}
	return "", nil
}

// return stdin, stdout, stderr
func redirectStd(g *Gossh) (err error) {

	// ??? what the fuck
	//fd := 0
	//_, err = terminal.MakeRaw(fd)
	//if err != nil {
	//	return errors.New(err.Error() + "make raw")
	//}

	// termWidth, termHeight affect to the output strings length, width
	// termType affect to the style of the
	termWidth, termHeight, termType := 5000, 50, "vt220"

	err = g.session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return errors.New(err.Error() + "RequestPty")
	}

	// redirect std
	g.stdin, err = g.session.StdinPipe()
	if err != nil {
		return errors.New(err.Error() + "StdinPipe")
	}

	g.stdout, err = g.session.StdoutPipe()
	if err != nil {
		return errors.New(err.Error() + "StdoutPipe")

	}

	g.stderr, err = g.session.StderrPipe()
	if err != nil {
		return errors.New(err.Error() + "StderrPipe")

	}

	err = g.session.Shell()
	if err != nil {
		return errors.New(err.Error() + "Shell")

	}

	return nil
}

func checkIfEnd(str string) bool {
	str = strings.TrimRight(str, " ")
	for _, end := range []byte{'#', ']'} {
		if strings.Index(str, string(end)) != -1 {
			return true
		}
	}
	return false
}
