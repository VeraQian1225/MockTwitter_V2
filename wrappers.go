package main

import (
	"log"
	"github.com/nzgogo/micro/codec"
	"github.com/nzgogo/micro/router"

)

func logMsgWrapper(handler router.Handler) router.Handler {
	return func(msg *codec.Message, reply string) *router.Error {
		log.Printf("[1] received message %v\n", *msg)
		err := handler(msg, reply)
		//log.Printf("[1] error message %v\n", err)
		return err
	}
}