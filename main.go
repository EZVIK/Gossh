package main

import (
	"fmt"
	"github.com/EZVIK/Gossh/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"time"
)

//var TERMINAL *service.SSHTerminal
var v *validator.Validate

var runtimeMap service.RuntimeMap

func init() {
	v = validator.New()
	runtimeMap = service.NewConnectMap()
	go runtimeMap.CheckClientTimeout()
}

func main() {
	config := fiber.Config{
		ReadTimeout: time.Second * 5,
	}

	app := fiber.New(config)

	app.Use(pprof.New())

	r := app.Group("/v1")
	r.Post("/run", Input)
	r.Post("/currConn", GetCurrConnMap)

	if err := app.Listen(":5588"); err != nil {
		fmt.Println("MAIN ERROR...", err)
		return
	}

	runtimeMap.CloseAll()
}

// Input & output
func Input(ctx *fiber.Ctx) error {
	n := service.CMD{}
	if err := BodyParse(ctx, &n); err != nil {
		return ctx.JSON(NewRes(err.Error()))
	}

	ans, err := runtimeMap.RunCmd(n.IP, n)

	if err != nil {
		return ctx.JSON(NewRes(err.Error()))
	}
	return ctx.JSON(NewRes(ans))
}

func GetCurrConnMap(ctx *fiber.Ctx) error {
	return ctx.JSON(NewRes(runtimeMap.GetConnList()))
}

// BodyParse http method

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

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
