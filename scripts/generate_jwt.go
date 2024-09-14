package main

import (
	mymiddleware "github.com/egreerdp/intern-project-template/internal/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	token, err := mymiddleware.GenerateJWT("Ewan")
	if err != nil {
		panic(err)
	}

	println(token)
}
