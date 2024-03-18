package manage

import (
	"log"
	"net/http"
	"package/platform/db"
	"package/platform/structs"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	c.HTML(http.StatusOK, "manage.html", nil)
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		return
	}

	var creator structs.Creator
	err := db.Db.Table("creators").Where("user = ?", user).First(&creator)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user is not a creator"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"authorized": true, "videos": getAllVideosFromACreator(creator)})

}

func getAllVideosFromACreator(creator structs.Creator) []structs.Video {
	// Fetch all the videos of a creator
	var videos []structs.Video
	db.Db.Table("videos").Where("owner = ?", creator).Find(&videos)
	log.Println("Retrieved all the videos of creator:", creator.CreatorName, videos)
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
