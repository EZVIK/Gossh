package main

import (
	"fmt"
	"github.com/EZVIK/Gossh/SSH_SIM"
	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const EndingArr = "#,~"

type SSHTerminal struct {
	Session    *ssh.Session //
	exitMsg    string       //
	stdout     io.Reader    //
	stdin      io.Writer    //
	stderr     io.Reader    //
	CurrStatus uint8        // status   0-init 1-replace 2-send_data
	CurrNode   SSH_SIM.Node //
}

func main() {

	var node_receiver = make(chan SSH_SIM.Node, 1000)
	var node_pusher = make(chan SSH_SIM.Node, 1000)

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("elish828MKB"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "159.75.82.148:22", sshConfig)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	err = New(client, node_receiver, node_pusher)
	if err != nil {
		fmt.Println(err)
	}
}

func (t *SSHTerminal) interactiveSession(receiver, pusher chan SSH_SIM.Node) error {

	// exit
	defer func() {
		if t.exitMsg == "" {
			fmt.Fprintln(os.Stdout, "the connection was closed on the remote side on ", time.Now().Format(time.RFC822))
		} else {
			fmt.Fprintln(os.Stdout, t.exitMsg)
		}
	}()

	// get std fd len
	fd := 0
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}

	defer terminal.Restore(fd, state)

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}

	termType := os.Getenv("TERM")
	//if termType == "" {
	//	termType = "xterm-256color"
	//}

	err = t.Session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	//t.updateTerminalSize()
	t.stdin, err = t.Session.StdinPipe()
	if err != nil {
		return err
	}
	t.stdout, err = t.Session.StdoutPipe()
	if err != nil {
		return err
	}
	t.stderr, err = t.Session.StderrPipe()

	// Stdin
	go t.Stdin(receiver)

	// Stdout
	go t.StdOut(pusher)

	// Stderr
	go t.StdErr(pusher)

	err = t.Session.Shell()
	if err != nil {
		return err
	}
	err = t.Session.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (t *SSHTerminal) check(err error) {
	if err != nil {
		fmt.Println("Stdin err:", err)
		t.exitMsg = err.Error()
		return
	}
}

func (t *SSHTerminal) Stdin(receiver chan SSH_SIM.Node) {
	//fmt.Println("Starting listening Receiver...")
	for {
		// TODO lock
		t.CurrNode = <-receiver
		fmt.Printf("receive...%s\n", t.CurrNode.Command)
		t.CurrStatus = 1 // 状态设置为等待数据
		_, err := t.stdin.Write(t.CurrNode.Command)
		if err != nil {
			fmt.Println("Stdin err:", err)
			t.exitMsg = err.Error()
			return
		}
	}
}

// StdOut push to channel
func (t *SSHTerminal) StdOut(Pusher chan SSH_SIM.Node) {

	sb := strings.Builder{}

	for {
		buf := make([]byte, 128)
		n, Outerr := t.stdout.Read(buf)
		if Outerr != nil {
			fmt.Println("StdOut err:", Outerr)
			return
		}

		if n > 0 {
			rawStr := string(buf[:n])
			sb.WriteString(rawStr)
			if checkIfEnd(rawStr) {
				t.CurrNode.Result = []string{sb.String()}
				Pusher <- t.CurrNode
				sb = strings.Builder{}
			}
		}
	}
}

func (t *SSHTerminal) StdErr(Pusher chan SSH_SIM.Node) {
	sb := strings.Builder{}
	buf := make([]byte, 128)
	last := "6961d7607f40a71bc7f0111a7c0bb443" // md5("last")
	for {
		n, Outerr := t.stderr.Read(buf)
		if Outerr != nil {
			fmt.Println("Stderr err:", Outerr)
			return
		}
		if n > 0 {
			rawStr := string(buf[:n])
			if rawStr != last {
				sb.Write(buf[:n])
				last = rawStr
			}
			if checkIfEnd(rawStr) {
				t.CurrNode.Result = []string{sb.String()}
				Pusher <- t.CurrNode
			}
		}
	}
}

func New(client *ssh.Client, receiver, pusher chan SSH_SIM.Node) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	// terminal
	s := SSHTerminal{
		Session: session,
	}

	// input
	go func() {
		time.Sleep(time.Second * 10)
		cmds := []string{"ls -l /home"}
		for _, cmd := range cmds {
			fmt.Println("push....", cmd)
			receiver <- NewNode([]byte(cmd))
		}
	}()

	// out
	go func() {
		for node := range pusher {
			fmt.Print(node.Result[0])
		}
	}()

	return s.interactiveSession(receiver, pusher)
}

// check if output end
func checkIfEnd(str string) bool {
	ends := strings.Split(EndingArr, ",")
	for _, end := range ends {
		if strings.Index(str, end) != -1 {
			return true
		}
	}
	return false
}

func GetSnowflakeId() string {
	//default node id eq 1,this can modify to different serverId node
	node, _ := snowflake.NewNode(10)
	// Generate a snowflake ID.
	id := node.Generate().String()
	return id
}

func NewNode(cmd []byte) SSH_SIM.Node {

	return SSH_SIM.Node{
		ID:      GetSnowflakeId(),
		Command: cmd,
	}
}

func (t *SSHTerminal) updateTerminalSize() {

	go func() {
		// SIGWINCH is sent to the process when the window size of the terminal has
		// changed.
		sigwinchCh := make(chan os.Signal, 1)
		signal.Notify(sigwinchCh, syscall.SIGWINCH)

		fd := 0
		termWidth, termHeight, err := terminal.GetSize(fd)
		if err != nil {
			fmt.Println(err)
		}

		for {
			select {
			// The client updated the size of the local PTY. This change needs to occur
			// on the server side PTY as well.
			case sigwinch := <-sigwinchCh:
				if sigwinch == nil {
					return
				}
				currTermWidth, currTermHeight, err := terminal.GetSize(fd)

				// Terminal size has not changed, don't do anything.
				if currTermHeight == termHeight && currTermWidth == termWidth {
					continue
				}

				t.Session.WindowChange(currTermHeight, currTermWidth)
				if err != nil {
					fmt.Printf("Unable to send window-change reqest: %s.", err)
					continue
				}

				termWidth, termHeight = currTermWidth, currTermHeight

			}
		}
	}()

}
