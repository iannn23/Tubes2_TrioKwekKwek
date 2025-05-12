package main

import (
	"log"
	"tubes2/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/elements", controller.GetElements)
	r.POST("/search", controller.SearchRecipe)

	log.Println("Server running at http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
