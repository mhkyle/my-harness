package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"mhkyle/my-harness/internal/route"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	r := gin.New()
	r.Use(gin.Recovery())

	route.Register(r)

	log.Printf("server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
