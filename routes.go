package main

import (
	"log"
	"github.com/nzgogo/micro/router"
	"MockTwitter_V2/controllers"
	"MockTwitter_V2/srv"
)

var routes []*router.Node

func init() {
	routes = append(routes, &router.Node{
		Method:  "Get",
		Path:    "/",
		ID:      "Show",
		Handler: user.Show,
	})
	routes = append(routes, &router.Node{
		Method:  "POST",
		Path:    "/create",
		ID:      "CreateUser",
		Handler: user.CreateUser,
	})
	routes = append(routes, &router.Node{
		Method:  "POST",
		Path:    "/login",
		ID:      "login",
		Handler: user.LoginWithCredential,
	})
	routes = append(routes, &router.Node{
		Method:  "Get",
		Path:    "/logout",
		ID:      "logout",
		Handler: user.UserLogOut,
	})
	routes = append(routes, &router.Node{
		Method:  "POST",
		Path:    "/post",
		ID:      "post",
		Handler: user.Post,
	})
}

func initRoutes() {
	log.Println("-----Service Routes List Start-----")
	for _, route := range routes {
		srv.Service.Options().Router.Add(route)
		log.Printf("%v\n", route)
	}
	log.Println("-----Service Routes List End-----")
}
