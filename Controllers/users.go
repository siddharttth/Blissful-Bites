package Controllers

import (
	DB "blissfulbites/DB"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UserDataHandler(c *gin.Context) {
	email := c.Query("email")
	data, err := DB.ReadRowData(email)
	if err != nil {
		fmt.Println("[User Data handler]", err)
	}
	trackJson, _ := DB.ReadTrack(email)

	if trackJson != nil {

		var nestedData []map[string]interface{}
		if err := json.Unmarshal(trackJson, &nestedData); err != nil {
			fmt.Println("[user details unmarshalling]", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal JSON data"})
			return
		}
		
		delete(data, "track")

		data["track"] = nestedData
	}
	c.JSON(http.StatusOK, data)
}

func AllUsersDataHandler(c *gin.Context) {
	allUsers, err := DB.ReadAllUsers("SELECT * FROM user_details")
	if err != nil {
		fmt.Println("[All users handler]", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
		return
	}
	c.HTML(http.StatusOK, "admin.html", allUsers)

}

func FirstLoginHandler(c *gin.Context) {
	email := c.Query("email")
	fmt.Println(email)
	exists, _ := DB.CheckEmailExists(email)
	if exists {
		c.JSON(http.StatusOK, gin.H{"message": "Email exists"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "Email does not exist"})
	}
}


