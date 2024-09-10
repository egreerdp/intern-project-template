package main

import (
	"log"

	"github.com/egreerdp/intern-project-template/api"
	"github.com/egreerdp/intern-project-template/db"
)

func main() {
	userStore := db.NewDatabase("user.db")

	handler := api.NewHandler(userStore)

	service := api.NewService(handler)

	log.Println("Init service")

	service.Start()
}
