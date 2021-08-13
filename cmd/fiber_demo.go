package main

import (
	"fmt"
	"github.com/EZVIK/Gossh/service"
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

//var TERMINAL *service.SSHTerminal
var v = validator.New()

var runtimeMap = service.NewConnectMap()

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

	n := service.CMD{}

	if err := BodyParse(ctx, &n); err != nil {
		return ctx.JSON(NewRes(err.Error()))
	}

	c, err := runtimeMap.Get(n.IP)

	cmds := strings.Split(n.Command, ";:;")

	//  TODO 2021/8/13 6:08 PM

	return ctx.JSON(NewRes(res))
	//return ctx.JSON(NewRes(ans))
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
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
//func (t *model.SSHTerminal) updateTerminalSize() {
//
//	go func() {
//		// SIGWINCH is sent to the process when the window size of the terminal has
//		// changed.
//		sigwinchCh := make(chan os.Signal, 1)
//		signal.Notify(sigwinchCh, syscall.SIGWINCH)
//
//		fd := int(os.Stdin.Fd())
//		termWidth, termHeight, err := terminal.GetSize(fd)
//		if err != nil {
//			fmt.Println(err)
//		}
//
//		for {
//			select {
//			// The client updated the size of the local PTY. This change needs to occur
//			// on the server side PTY as well.
//			case sigwinch := <-sigwinchCh:
//				if sigwinch == nil {
//					return
//				}
//				currTermWidth, currTermHeight, err := terminal.GetSize(fd)
//
//				// Terminal size has not changed, don't do anything.
//				if currTermHeight == termHeight && currTermWidth == termWidth {
//					continue
//				}
//
//				t.Session.WindowChange(currTermHeight, currTermWidth)
//				if err != nil {
//					fmt.Printf("Unable to send window-change reqest: %s.", err)
//					continue
//				}
//
//				termWidth, termHeight = currTermWidth, currTermHeight
//
//			}
//		}
//	}()
//
//}