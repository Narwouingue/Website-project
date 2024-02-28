package CreatorPrivatePage

import (
	"log"
	"net/http"
	"package/platform/db"
	"package/platform/structs"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	userName := c.Param("userName")
	var user structs.User

	err := db.Db.Table("users").Where("username = ?", userName).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Creator not found"})
		return
	}
	if !user.IsCreator {
		c.JSON(http.StatusNotFound, gin.H{"error": "This user is not a creator"})
		return
	}

	connectedUser, message := GetConnectedUser(c)

	err = db.Db.Table("users").Where("username = ?", userName).Where("subcribers = ?", connectedUser).Find(&user).Error
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "You're not subscribed to this creator"})
		return
	}
	videos := getAllVideosFromACreator(userName)
	var privateVideos []structs.Video
	for _, v := range videos {
		if !v.IsPublic {
			privateVideos = append(privateVideos, v)
		}
	}

	c.JSON(http.StatusAccepted, gin.H{"videos": privateVideos, "message": message})
}

func getAllVideosFromACreator(username string) []structs.Video {
	var videos []structs.Video
	db.Db.Table("videos").Where("owner = ?", username).Find(&videos)
	log.Println(videos)
	return videos

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
