package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Add your frontend URL
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Accept", "api-key", "Content-Type"},
	}))

	r.GET("/points", s.getAllPointsHandler)
	r.POST("/points/add", s.addPointsHandler)
	r.GET("/points/:username", s.getPointsHandler)
	r.POST("/points/remove", s.removePointsHandler)

	return r
}

func (s *Server) getAllPointsHandler(c *gin.Context) {
	users, err := s.db.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *Server) addPointsHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Points   int    `json:"points"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAmount, err := s.db.AddPoints(req.Username, req.Points)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"newAmount": newAmount})
}

func (s *Server) removePointsHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Points   int    `json:"points"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success, newAmount, err := s.db.RemovePoints(req.Username, req.Points)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if success {
		c.JSON(http.StatusOK, gin.H{"success": true, "newAmount": newAmount})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false})
	}
}

func (s *Server) getPointsHandler(c *gin.Context) {
	username := c.Param("username")

	points, err := s.db.GetPoints(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"points": points})
}
