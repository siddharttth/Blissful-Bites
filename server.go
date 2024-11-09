// main.go

package main

import (
	AI "blissfulbites/AI"
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
	err := godotenv.Load("cred.env")
	if err != nil {
		fmt.Println("[server] Error loading .env file")
	}

	AI.InitializeModel(os.Getenv("GOOGLE_API_KEY"))

	_ = DB.ConnectPsql(os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASS"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"))
	_ = DB.CreateTableUserDetails()

	r := gin.Default()

	// r.LoadHTMLGlob("static/client/*")
	r.LoadHTMLGlob("static/*.html")
	// r.Static("/css", "static/css")
	// r.Static("/logos", "static/logos")
	r.Static("/static", "./static")
	// r.Static("/static/js", "./static/js")
	r.Static("/images", "./static/images")
	r.Static("/intlTelInput", "./static/intlTelInput")

	//client endpoints
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
		Controllers.UserDataHandler(c)
	})

	r.GET("/firstlogin", func(c *gin.Context) {
		Controllers.FirstLoginHandler(c)
	})

	r.POST("/genDietPlan", func(c *gin.Context) {
		Controllers.GenDietPlan(c)
	})

	// admin endpoints

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

	// r.Static("/static", "./static")
	// r.Run(":8000")
	log.Fatal(r.Run(":3000"))
}
