package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/nzgogo/micro"
	"github.com/nzgogo/micro/api"
	"github.com/nzgogo/micro/codec"
	"github.com/nzgogo/micro/constant"
	"github.com/nzgogo/micro/context"
	recpro "github.com/nzgogo/micro/recover"
)

type MyHandler struct {
	srv gogo.Service
}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if recMsg := recover(); recMsg != nil {
			recpro.Recover(h.srv.Options().Transport.Options().Subject, "ServeHTTP", recMsg, r)
		}
	}()
	config := h.srv.Options()

	ctxId := config.Context.Add(&context.Conversation{
		Response: w,
	})
	// map the HTTP request to internal transport request message struct.
	request, err := gogoapi.HTTPReqToIntrlSReq(r, config.Transport.Options().Subject, ctxId)
	if err != nil {
		http.Error(w, "Cannot process request", http.StatusNotFound)
		return
	}

	//look up registered service in kv store
	err = config.Router.HttpMatch(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	srvName := gogo.URLToServiceName(request.Host, request.Path)
	srvVersion := gogo.URLToServiceVersion(request.Path)
	log.Println("Parsed service: " + srvName + "-" + srvVersion)

	//service discovery

	subj, err := config.Selector.Select(srvName, srvVersion)
	if err != nil {
		http.Error(w, "Cannot process request", http.StatusInternalServerError)
		panic("None service available. " + err.Error())
	}
	log.Println("Found service: " + subj)

	//transport file
	reqBody := make(map[string]interface{})
	codec.Unmarshal(request.Body, &reqBody)
	if reqBody["file"] != nil {
		log.Println("File found!")
		fileBytes, ok := reqBody["file"].(string)
		if !ok {
			http.Error(w, "Cannot process request", http.StatusInternalServerError)
			panic("Failed to parse file from request body. ")
		}
		if fileSub, err := config.Selector.Select(constant.FILE_SERVICE_NAME, constant.FILE_SERVICE_VERSION); err != nil {
			http.Error(w, "Cannot process request", http.StatusInternalServerError)
			panic("None service available. " + err.Error())
		} else {
			h.srv.Options().Transport.SendFile(request, fileSub, fileBytes)
			log.Println("Send file to transport.")
		}
		delete(reqBody, "file")
		requestBodyWithoutFile, _ := codec.Marshal(reqBody)
		request.Body = requestBodyWithoutFile
	}

	// transport request
	bytes, _ := codec.Marshal(request)
	log.Println("Send to service: " + subj)
	respErr := config.Transport.Publish(subj, bytes)

	if respErr != nil {
		http.Error(w, "Transport error", http.StatusInternalServerError)
		panic("Transport error, Published failed. " + respErr.Error())
	}

	config.Context.Wait(ctxId)
}

func main() {
	service := gogo.NewService(
		"gogo-core-api",
		"v1",
	)
	service.Options().Transport.SetHandler(service.ApiHandler)

	var respWrapChain = []gogo.HttpResponseWrapper{
		logHttpRespWrapper,
	}

	if err := service.Init(gogo.WrapRepsWriter(respWrapChain...)); err != nil {

		log.Fatal(err)
	}

	go func() {
		if err := service.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	handler := MyHandler{service}
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: &handler,
	}

	// Http wrapper
	var httpChain = []gogo.HttpHandlerWrapper{
		logWrapper,
	}
	http.HandleFunc("/", gogo.HttpWrapperChain(handler.ServeHTTP, httpChain...))

	go func() {
		if err := server.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	select {
	// wait on kill signal
	case <-ch:
		service.Stop()
		if err := server.Shutdown(nil); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	}
}
