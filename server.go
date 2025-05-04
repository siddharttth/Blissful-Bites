package main

import (
	AI "blissfulbites/AI"
	Auth "blissfulbites/Auth"
	Controllers "blissfulbites/Controllers"
	DB "blissfulbites/DB"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from cred.env
	fmt.Println("🔍 Loading environment variables...")
	err := godotenv.Load("cred.env")
	if err != nil {
		log.Fatalf("❌ Error loading cred.env file: %s", err)
	} else {
		fmt.Println("✅ Environment variables loaded successfully.")
	}

	// Retrieve and display the port being used
	port := os.Getenv("PORT")
	fmt.Printf("🌟 Using PORT: %s\n", port)

	// Initialize Groq model
	groqApiKey := os.Getenv("GROQ_API_KEY")
	fmt.Printf("🔑 Using Groq API key: %s\n", groqApiKey)
	AI.InitializeGroqModel(groqApiKey)
	fmt.Println("✨ Groq model initialized.")

	// Connect to PostgreSQL database
	fmt.Println("🗄️  Connecting to PostgreSQL database...")
	err = DB.ConnectPsql(os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASS"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"))
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL database: %s", err)
	} else {
		fmt.Println("✅ Connected to PostgreSQL database successfully.")
	}

	// Run database migrations
	fmt.Println("🔄 Running database migrations...")
	err = DB.MigrateDB()
	if err != nil {
		log.Fatalf("❌ Failed to run database migrations: %s", err)
	}

	// Set up Gin server
	fmt.Println("⚙️  Setting up server...")
	r := gin.Default()

	// Set trusted proxies to localhost for security
	r.SetTrustedProxies([]string{"127.0.0.1"})
	fmt.Println("🔐 Trusted proxies set to localhost.")

	// Switch to release mode for production
	gin.SetMode(gin.ReleaseMode)
	fmt.Println("🚀 Running in release mode...")

	// Load HTML templates and static files
	r.LoadHTMLGlob("static/*.html")
	r.Static("/static", "./static")
	r.Static("/images", "./static/images")
	r.Static("/intlTelInput", "./static/intlTelInput")

	// Initialize DBAuth instance
	auth := Auth.NewDBAuth()

	// Define client endpoints
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth.html", gin.H{})
	})

	r.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", gin.H{})
	})

	r.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "Home.html", gin.H{})
	})

	r.GET("/form", func(c *gin.Context) {
		c.HTML(http.StatusOK, "form.html", gin.H{})
	})

	r.GET("/track", func(c *gin.Context) {
		c.HTML(http.StatusOK, "track.html", gin.H{})
	})

	r.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "Contact.html", gin.H{})
	})

	r.POST("/contactUs", func(c *gin.Context) {
		Controllers.ContactHandler(c)
	})

	r.POST("/userFormDetails", func(c *gin.Context) {
		Controllers.FormHandler(c)
	})

	r.POST("/trackMeal", func(c *gin.Context) {
		Controllers.AppendMealsHandler(c)
	})

	r.GET("/userDetails", func(c *gin.Context) {
		Controllers.FormUserDataHandler(c)
	})

	r.GET("/userBasicInfo", func(c *gin.Context) {
		Controllers.GetUserBasicInfo(c)
	})

	r.GET("/userBMI", func(c *gin.Context) {
		Controllers.GetUserBMI(c)
	})

	r.GET("/userHealthScore", func(c *gin.Context) {
		Controllers.GetUserHealthScore(c)
	})

	r.GET("/firstlogin", func(c *gin.Context) {
		Controllers.FirstLoginHandler(c)
	})

	r.POST("/genDietPlan", func(c *gin.Context) {
		Controllers.GenDietPlan(c)
	})

	// Define admin endpoints
	r.GET("/admin", func(c *gin.Context) {
		Controllers.AllUsersDataHandler(c)
	})

	r.GET("/dm", func(c *gin.Context) {
		Controllers.DmHandler(c)
	})

	r.GET("/user", func(c *gin.Context) {
		c.HTML(http.StatusOK, "user.html", gin.H{})
	})

	r.POST("/updateDiet", func(c *gin.Context) {
		Controllers.UpdateDietHandler(c)
	})

	// Add POST route for signup
	r.POST("/signup", func(c *gin.Context) {
		var json struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		success := auth.Signup(json.Username, json.Password)
		if success {
			c.JSON(http.StatusOK, gin.H{"message": "Signup successful"})
		} else {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists or error occurred"})
		}
	})

	// Add POST route for signin
	r.POST("/signin", func(c *gin.Context) {
		var json struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		success := auth.Login(json.Username, json.Password)
		if success {
			c.JSON(http.StatusOK, gin.H{"message": "Signin successful"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		}
	})

	// Handle favicon request
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.String(http.StatusNoContent, "")
	})

	// Start the server
	fmt.Printf("🌍 Starting server on http://localhost:%s...\n", port)
	log.Fatal(r.Run("localhost:" + port))
}
