package main
import (
	"log"
	"net/http"
	"github.com/nzgogo/micro"
	"github.com/nzgogo/micro/codec"
	"github.com/nzgogo/micro/router"
)

func logMsgWrapper(handler router.Handler) router.Handler {
	return func(msg *codec.Message, reply string) *router.Error {
		log.Printf("[1] received message %v\n", *msg)
		if msg.Body != nil {

		}
		err := handler(msg, reply)
		//log.Printf("[1] error message %v\n", err)
		return err
	}
}

func logHttpRespWrapper(writeResponse gogo.HttpResponseWriter) gogo.HttpResponseWriter  {
	return func(rw http.ResponseWriter, response *codec.Message) {
		//log.Printf("[1] logHttpRespwrapper -> message to send to http ResponseWriter: %s\n" , response.Body)
		log.Printf("[1] logHttpRespwrapper before")
		writeResponse(rw, response)
		log.Printf("[1] logHttpRespwrapper after")
	}
}

func logWrapper(wrapper gogo.HttpHandlerFunc) gogo.HttpHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("[1] HttpHandlerWrapper before")
		wrapper(w,r)
		log.Println("[1] HttpHandlerWrapper after")
	}
}
