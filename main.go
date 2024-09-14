package main

import (
	"log"

	"github.com/egreerdp/intern-project-template/api"
	"github.com/egreerdp/intern-project-template/db"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	userStore := db.NewDatabase("user.db")

	handler := api.NewHandler(userStore)

	service := api.NewService(handler)

	log.Println("Init service")

	service.Start()
}
