package user

import (
	"log"
	"net/http"

	"package/platform/db"
	"package/platform/structs"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {

	connectedUser, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not connected"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"fullName": connectedUser.FullName, "username": connectedUser.UserName, "email": connectedUser.Email, "tokens": connectedUser.Tokens})
	if connectedUser.IsCreator {
		videos := getAllVideosFromACreator(connectedUser.UserName)
		c.JSON(http.StatusAccepted, gin.H{"videos": videos})
	}

}

func GetConnectedUser(c *gin.Context) (returnedUser structs.User, message string) {
	token, err := c.Cookie("connection")

	if err != nil {
		var user structs.User
		return user, "error"
	}
	var user structs.User
	db.Db.Table("users").Where("token = ?", token).First(&user)
	return user, "success"

}

func getAllVideosFromACreator(username string) []structs.Video {
	var videos []structs.Video
	db.Db.Table("videos").Where("owner = ?", username).Find(&videos)
	log.Println(videos)
	return videos

}
