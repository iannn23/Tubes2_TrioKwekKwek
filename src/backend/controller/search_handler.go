package controller

import (
	"net/http"
	"tubes2/algorithm"
	"tubes2/model"

	"github.com/gin-gonic/gin"
)

func GetElements(c *gin.Context) {
	elements, err := model.LoadElements("data/elements.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, elements)
}

type SearchRequest struct {
	Target    string `json:"target"`
	Algorithm string `json:"algorithm"` // "BFS" or "DFS"
	Multiple  bool   `json:"multiple"`
	MaxRecipe int    `json:"maxRecipe"`
}

func SearchRecipe(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	elements, err := model.LoadElements("data/elements.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := algorithm.Search(req.Target, req.Algorithm, req.Multiple, req.MaxRecipe, elements)
	c.JSON(http.StatusOK, result)
}
