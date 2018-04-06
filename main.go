package main

import (
	"log"
	gogo "github.com/nzgogo/micro"
	"MockTwitter_V2/srv"
)

func main(){
	defer srv.UserDB.Close()

	if err := srv.Service.Init(gogo.WrapHandler(logMsgWrapper)); err != nil {
		log.Fatal(err)
	}

	srv.Service.Options().Transport.SetHandler(srv.Service.ServerHandler)
	initRoutes()

	// Run server
	if err := srv.Service.Run(); err != nil {
		log.Fatal(err)
	}

}
