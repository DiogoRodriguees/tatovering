package routes

import (
	"context"
	"net/http"
	"tatovering/src/controllers"
	"tatovering/src/middlewares"

	"github.com/gin-gonic/gin"
	supabase "github.com/nedpals/supabase-go"
)

func SetupUsuariosRoutes(router *gin.Engine, client *supabase.Client) {
	usuarioGroup := router.Group("/usuarios")
	{
		usuarioGroup.POST(
			"",
			middlewares.JwtTokenCheck(client),
			controllers.CadastrarUsuario(client),
		)
		usuarioGroup.GET(
			"/:id",
			controllers.GetById(client),
		)
		usuarioGroup.PATCH(
			"/:id",
			middlewares.JwtTokenCheck(client),
			controllers.EditarUsuario(client),
		)
		usuarioGroup.DELETE(
			"/:id",
			middlewares.JwtTokenCheck(client),
			controllers.DeletarUsuario(client),
		)

		router.POST("/signup", SignUp(client))
		router.POST("/signin", SignIn(client))

	}
}

func SignUp(client *supabase.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Defines the input data and validation
		var requestBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		// Bind the request to the defined model and throw error if some validation fails.
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create user credentials
		credentials := supabase.UserCredentials{
			Email:    requestBody.Email,
			Password: requestBody.Password,
		}
		ctx := context.Background()
		// Sign up the user with Supabase
		user, err := client.Auth.SignUp(ctx, credentials)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"user": user})
	}
}

func SignIn(client *supabase.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create user credentials
		credentials := supabase.UserCredentials{
			Email:    requestBody.Email,
			Password: requestBody.Password,
		}

		ctx := context.Background()
		// Sign up the user with Supabase
		user, err := client.Auth.SignIn(ctx, credentials)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}
