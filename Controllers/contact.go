package Controllers

import (
	DB "blissfulbites/DB"
	// "encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ContactHandler(c *gin.Context) {
	email := c.PostForm("email")
	message := c.PostForm("message")

	err := DB.UpdateDM(email, message)
	if err != nil {
		fmt.Println("[Contact handler]",err)
		c.JSON(http.StatusInternalServerError, gin.H{"status":"message couldn't be sent"})
	}

	c.JSON(http.StatusOK, gin.H{"status":"message sent"})

}

func DmHandler(c *gin.Context){
	allDms, err := DB.ReadAllUsers("SELECT name,email, dm FROM user_details")
	if err != nil {
		fmt.Println("[All Dms handler]", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
		return
	}
	// fmt.Println(allDms)
	c.HTML(http.StatusOK, "dm.html", allDms)
}
