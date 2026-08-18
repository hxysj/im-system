package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H struct {
	Code  int  `json:"code"`
	Msg   string `json:"msg"`
	Data  interface{} `json:"data"`
	Total interface{} `json:"total"`
}


func Resp(w http.ResponseWriter,code int,data interface{},msg string) {
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code:code,
		Data:data,
		Msg:msg,
	}

	ret,err := json.Marshal(h)
	if err != nil{
		fmt.Println(err)
	}
	w.Write(ret)
}

func RespList(w http.ResponseWriter,code int, data interface{},total interface{}){
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code: code,
		Data: data,
		Total: total,
	}

	ret,err := json.Marshal(h)
	if err != nil{
		fmt.Println(err)
	}
	w.Write(ret)
}

func RespFail(w http.ResponseWriter,msg string) {
	Resp(w,-1,nil,msg)
}

func RespOk(w http.ResponseWriter,data interface{},msg string) {
	Resp(w,0,data,msg)
}

func RespOkList(w http.ResponseWriter,data interface{},total interface{}){
	RespList(w,0,data,total)
}
