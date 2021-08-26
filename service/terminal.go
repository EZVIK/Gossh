package service

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"strings"
	"time"
)

type Terminal struct {
	session *ssh.Session
	exitMsg string
	stdout  io.Reader
	stdin   io.Writer
	stderr  io.Reader
	device  *SSHDevice
}

// init ssh Session setting
func (t *Terminal) interactiveSession(device *SSHDevice) error {

	fd := 0
	_, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}

	// termWidth, termHeight affect to the output strings length, width
	// termType affect to the style of the
	termWidth, termHeight, termType := 500, 30, "vt220"

	//
	err = t.session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	// redirect std
	t.stdin, err = t.session.StdinPipe()
	if err != nil {
		return err
	}
	t.stdout, err = t.session.StdoutPipe()
	if err != nil {
		return err
	}
	t.stderr, err = t.session.StderrPipe()

	err = t.session.Shell()
	if err != nil {
		return err
	}

	// filtering login msg
	// if loginLog is false, the login msg will send with the first response
	if device.loginLog {
		t.Read()
	}

	return nil
}

// Run cmd return Output
func (t *Terminal) Run(commands []string) (ans map[string][]string, err error) {

	// data result set
	ans = make(map[string][]string, 0)

	// input & output
	for _, cmd := range commands {

		// write cmd
		if err = t.Write(cmd); err != nil {
			ans[cmd] = []string{}
		}

		sb := t.Read()

		// split with enter
		tmp := strings.Split(sb, "\r\n")

		// delete prefix & command
		tmp = tmp[1 : len(tmp)-2]

		// append to result arr
		ans[cmd] = tmp
	}

	return
}

// Write command in stdin
func (t *Terminal) Write(cmd string) error {

	//  \r\n simulate the "enter"
	_, err := t.stdin.Write([]byte(cmd + "\r\n"))

	return err
}

// Read for Std
func (t *Terminal) Read() string {
	timer1 := time.NewTicker(t.device.timeout)
	ans := ""

	for {
		select {
		// Timeout timer, break loop when timer is <-
		case <-timer1.C:
			return ans
		// continue read from std
		default:
			read, err := t.readOnce()
			if err != nil {
				log.Fatal(err)
			}
			ans += read

			// check stdout finished output
			if t.device.checkIfEnd(read) {
				return ans
			}
		}
	}
}

// ReadFrom stdout once
func (t *Terminal) readOnce() (string, error) {
	ans := ""
	buf := make([]byte, 128)
	n, stdoutErr := t.stdout.Read(buf)
	if stdoutErr != nil {
		return ans, stdoutErr
	}

	if n > 0 {
		rawStr := string(buf[:n])
		ans = ans + fmt.Sprintf("%s", rawStr)
		return ans, nil
	}

	return "", nil
}
