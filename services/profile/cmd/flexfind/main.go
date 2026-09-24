package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
Что должен делать:

- профиль кандидата;

- skills;

- preferred positions;

- location;

- experience;

- education;

- job preferences.
*/

func main() {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, I am a flexfind")
	})

	server := &http.Server{
		Addr: ":8080",
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}