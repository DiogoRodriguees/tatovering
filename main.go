package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	supabase "github.com/nedpals/supabase-go"
)

func main() {
	supabaseURL := "https://pfzlboeaonsookzcnniv.supabase.co"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InBmemxib2Vhb25zb29remNubml2Iiwicm9sZSI6ImFub24iLCJpYXQiOjE2OTIwNjI3NTYsImV4cCI6MjAwNzYzODc1Nn0.KuEEX9EBIQmLTA02iPtqqNIewDmXITDxnIfD4qEqTN8"

	// Create a Gin router
	router := gin.Default()

	//Initialize a single supabase client instead of one for each query received
	client := supabase.CreateClient(supabaseURL, supabaseKey)

	extractBearerToken := func(header string) (string, error) {
		if header == "" {
			return "", errors.New("Missing authorization header")
		}

		jwtToken := strings.Split(header, " ")
		if len(jwtToken) != 2 {
			return "", errors.New("Incorrectly formatted authorization header")
		}

		return jwtToken[1], nil
	}

	jwtTokenCheck := func(c *gin.Context) {
		jwtToken, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		client.DB.AddHeader("Authorization", "Bearer "+jwtToken)
		c.Next()
	}

	// Create a group, all routes initialized with this group will pass through the
	// jwtTokenCheck middleware function and be located like: /private/...
	private := router.Group("/", jwtTokenCheck)

	// Route for user sign-up
	router.POST("/signup", func(c *gin.Context) {
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
	})

	// Route for user sign-in
	router.POST("/signin", func(c *gin.Context) {
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
	})

	// Define CRUD routes for "usuarios"
	private.POST("/usuarios", func(c *gin.Context) {
		// Create a new usuario
		var row Usuario

		if errBind := c.ShouldBindJSON(&row); errBind != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errBind.Error()})
			return
		}

		var results []Usuario
		errInsert := client.DB.From("usuarios").Insert(row).Execute(&results)

		if errInsert != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errInsert.Error()})
			return
		}

		c.JSON(http.StatusCreated, results)
	})

	router.GET("/usuarios/:id", func(c *gin.Context) {
		id := c.Param("id")
		var user Usuario
		err := client.DB.From("usuarios").Select("*").Single().Eq("id", id).Execute(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	private.PATCH("/usuarios/:id", func(c *gin.Context) {
		id := c.Param("id")
		var user Usuario
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var results []Usuario
		err := client.DB.From("usuarios").Update(user).Eq("id", id).Execute(&results)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	/*********************************************************
	* 				   	  CRUD TATUAGENS 				   	 *
	**********************************************************/
	private.POST("/tatuagens", func(c *gin.Context) {

		var requestBody Tatuagem

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var results []Tatuagem

		// inserting data and receive error if exist
		err := client.DB.From("tatuagens").Insert(requestBody).Execute(&results)

		// chack error returned
		if err != nil {
			// ginh.H used to returnd a json file
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// return a json file
		c.JSON(http.StatusOK, results)

	})

	private.PATCH("/tatuagens/:id", func(c *gin.Context) {
		tatuagemId := c.Param("id")

		var requestBody Tatuagem

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var results Tatuagem
		err := client.DB.From("tatuagens").Select("*").Single().Eq("id", tatuagemId).Execute(&results)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updateErr := client.DB.From("tatuagens").Update(requestBody).Eq("id", tatuagemId)

		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errror": err.Error()})
		}

		c.JSON(http.StatusOK, results)
	})

	// Find all tattoo by tattoo artist
	router.GET("/tatuagens/:id", func(c *gin.Context) {
		// extract of param the tatuador id
		tatuadorId := c.Param("id")

		// variable of return function execute databse
		var results []Tatuagem

		err := client.DB.From("tatuagens").Select("*").Eq("tatuador_id", tatuadorId).Execute(&results)

		// tratament error case exists
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// response
		c.JSON(http.StatusOK, results)
	})

	// Delete tattoo per id tattoo artist
	private.DELETE("/tatuagens/:id", func(c *gin.Context) {
		tatuagemId := c.Param("id")

		var results Tatuagem
		err := client.DB.From("tatuagens").Delete().Eq("id", tatuagemId).Execute(&results)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
	})

	// Start the Gin server
	port := 8080 // Change to the desired port
	router.Run(fmt.Sprintf(":%d", port))
}

// Define the Usuario struct to match your database structure
type Usuario struct {
	Nome string `json:"nome"`
}

type Tatuagem struct {
	TatuadorId    int     `json:"tatuador_id"`
	AgendamentoId int     `json:"agendamento_id"`
	Preco         float32 `json:"preco"`
	Desenho       string  `json:"desenho"`
	Tamaho        int     `json:"tamanho"`
	Cor           string  `json:"cor"`
	Estilo        string  `json:"estilo"`
}

// Helper function to convert Usuario struct to map for Supabase
func tatuadorToMap(usuario Usuario) map[string]interface{} {
	return map[string]interface{}{
		"nome": usuario.Nome,
	}
}
