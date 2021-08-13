package main

import (
	"fmt"
	"github.com/EZVIK/Gossh/SSH_SIM"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type SSHTerminal struct {
	Session *ssh.Session
	exitMsg string
	stdout  io.Reader
	stdin   io.Writer
	stderr  io.Reader
}

type CMD struct {
	Namespace string `validate:"required" json:"namespace"`
	IP        string `validate:"required" json:"ip"`
	Command   string `validate:"required" json:"cmd"`
}

type ConnectMap map[string]SSHClient

type SSHClient struct {
	ID        string
	client    *ssh.Client
	terminals *SSHTerminal
	settings  SSHDevice
}

type SSHDevice struct {
	Host       string
	Port       uint16
	username   string
	password   string
	termType   string
	termHeight int
	termWidth  int
	loginLog   bool
	timeout    int // second
}

func (s *SSHDevice) GetUsername() string {
	return s.username
}

func (s *SSHDevice) GetPassword() string {
	return s.password
}

const EndingArr = "#"
const host = "43.128.63.180" // 43.128.63.180:22  159.75.82.148
const pass = "elish000MKB"

var CONNMAP = make(map[string]*SSHClient, 0)

//var TERMINAL *SSHTerminal
var pusher = make(chan SSH_SIM.Node, 1000)
var v = validator.New()
var deviceDatabase = []SSHDevice{
	{"159.75.82.148", 22, "root", "elish828MKB", "vt220", 500, 40, true, 300},
	{"43.128.63.180", 22, "root", "elish000MKB", "vt220", 500, 40, true, 300},
}

func main() {

	app := fiber.New()
	Router(app)

	if err := app.Listen(":5588"); err != nil {
		fmt.Println("MAIN ERROR...", err)
		return
	}
}

func Router(app *fiber.App) {
	app.Use(cors.New())

	app.Post("/run", Input)
}

// Input & output
func Input(ctx *fiber.Ctx) error {

	n := CMD{}

	err := BodyParse(ctx, &n)
	var device SSHDevice
	for _, d := range deviceDatabase {
		if d.Host == n.IP {
			device = d
		}
	}

	var rawSC = new(SSHClient)

	if sc, ok := CONNMAP[device.Host]; !ok {
		rawSC.Login(device)
	} else {
		rawSC = sc
	}

	cmds := strings.Split(n.Command, ";:;")

	sb := ""
	buf := make([]byte, 128)
	res := make([]string, 0)

	for _, cmd := range cmds {
		_, err = rawSC.terminals.stdin.Write([]byte(cmd + "\r\n"))
		if err != nil {
			fmt.Println(err)
			rawSC.terminals.exitMsg = err.Error()
			return ctx.JSON(rawSC.terminals.exitMsg)
		}

		for {
			n, Outerr := rawSC.terminals.stdout.Read(buf)
			if Outerr != nil {
				fmt.Println("StdOut err:", Outerr)
				return ctx.JSON(NewRes(rawSC.terminals.exitMsg))
			}

			if n > 0 {
				rawStr := string(buf[:n])
				sb = sb + fmt.Sprintf("%s", rawStr)
				if checkIfEnd(rawStr) {
					break
				}
			}
		}

		ans := strings.Split(sb, "\r\n")
		ans = ans[1 : len(ans)-2]
		res = append(res, ans...)
	}

	//ans, _ := strconv.Unquote(sb)
	//fmt.Println(sb)
	return ctx.JSON(NewRes(res))
	//return ctx.JSON(NewRes(ans))
}

// init ssh Session setting
func (t *SSHTerminal) interactiveSession(device SSHDevice) error {

	// close
	defer func() {
		if t.exitMsg == "" {
			fmt.Fprintln(os.Stdout, "the connection was closed on the remote side on ", time.Now().Format(time.RFC822))
		} else {
			fmt.Fprintln(os.Stdout, t.exitMsg)
		}
	}()

	fd := 0
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer terminal.Restore(fd, state)

	//termWidth, termHeight, err := terminal.GetSize(fd)
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
			if checkIfEnd(strr) {
				fmt.Println(sb)
				break
			}
		}
	}

	go t.Session.Wait()

	return nil
}

// New Session
func (s *SSHClient) New() error {

	session, err := s.client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	s.terminals = &SSHTerminal{
		Session: session,
	}

	return s.terminals.interactiveSession(s.settings)
}

func (s *SSHClient) Login(dev SSHDevice) (err error) {

	CONNMAP[dev.Host] = s

	sshConfig := &ssh.ClientConfig{
		User: dev.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(dev.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 300,
	}

	// connect & save client
	c, err := ssh.Dial("tcp", dev.Host+":"+strconv.Itoa(int(dev.Port)), sshConfig)
	if err != nil {
		return err
	}
	s.client = c

	return s.New()
}

func (d *SSHDevice) InitClient() {

}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// check if output end
func checkIfEnd(str string) bool {
	str = strings.TrimRight(str, " ")
	ends := strings.Split(EndingArr, ",")
	for _, end := range ends {
		if strings.Index(str, end) != -1 {
			return true
		}
	}
	return false
}

// BodyParse http method

func BodyParse(ctx *fiber.Ctx, dto interface{}) error {
	_ = ctx.BodyParser(dto)        // 解析参数
	validateError := v.Struct(dto) // 校验参数
	if validateError != nil {
		return validateError
	}
	return nil
}

func NewRes(data interface{}) Response {
	return Response{Code: 200, Msg: "Success", Data: data}
}

// --- useless
func (t *SSHTerminal) updateTerminalSize() {

	go func() {
		// SIGWINCH is sent to the process when the window size of the terminal has
		// changed.
		sigwinchCh := make(chan os.Signal, 1)
		signal.Notify(sigwinchCh, syscall.SIGWINCH)

		fd := int(os.Stdin.Fd())
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
