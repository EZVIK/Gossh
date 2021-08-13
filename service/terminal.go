package service

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"strings"
	"time"
)

type Shell struct {
	Session *ssh.Session
	exitMsg string
	stdout  io.Reader
	stdin   io.Writer
	stderr  io.Reader
	timeout time.Duration
	device  *SSHDevice
}

// init ssh Session setting
func (t Shell) interactiveSession(device SSHDevice) error {

	fd := 0
	_, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}

	termWidth, termHeight := 500, 30

	if err != nil {
		return err
	}

	termType := "vt220"

	err = t.Session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	//t.updateTerminalSize()

	// redirect
	t.stdin, err = t.Session.StdinPipe()
	if err != nil {
		return err
	}
	t.stdout, err = t.Session.StdoutPipe()
	if err != nil {
		return err
	}
	t.stderr, err = t.Session.StderrPipe()

	err = t.Session.Shell()
	if err != nil {
		return err
	}

	// filtering login msg
	if device.loginLog {
		sb := ""
		for {
			buf := make([]byte, 128)
			n, Outerr := t.stdout.Read(buf)
			if Outerr != nil {
				fmt.Println("StdOut err:", Outerr)
			}
			//Last failed login: Thu Aug 12 15:58:51 CST 2021 from 222.187.232.205 on ssh:notty\r\n There were 2 failed login attempts since the
			strr := string(buf[:n])
			sb = sb + strr

			if device.checkIfEnd(strr) {
				fmt.Println(sb)
				break
			}
		}
	}

	//go t.Session.Wait()

	return nil
}

// Run cmd return Output
func (t Shell) Run(cmd CMD) (ans []string, err error) {

	// input & output
	sb := ""
	buf := make([]byte, 128)
	_, err = t.stdin.Write([]byte(cmd.Command + "\r\n"))
	if err != nil {
		fmt.Println(err)
		//t.exitMsg = err.Error()
		return
	}

	for {
		n, Outerr := t.stdout.Read(buf)
		if Outerr != nil {
			fmt.Println("StdOut err:", Outerr)

			return
		}

		if n > 0 {
			rawStr := string(buf[:n])
			sb = sb + fmt.Sprintf("%s", rawStr)
			if t.device.checkIfEnd(rawStr) {
				break
			}
		}
	}

	ans = strings.Split(sb, "\r\n")
	ans = ans[1 : len(ans)-2]

	return
}
