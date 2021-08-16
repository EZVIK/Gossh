package main

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func NewRes(data interface{}) Response {
	return Response{Code: 200, Msg: "Success", Data: data}
}
