package creatorPage

import (
	"log"
	"net/http"
	"package/db"
	"package/structs"

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

	videos := getAllVideosFromACreator(userName)

	c.JSON(http.StatusAccepted, gin.H{"videos": videos})
}

func getAllVideosFromACreator(username string) []structs.Video {
	var videos []structs.Video
	db.Db.Table("videos").Where("owner = ?", username).Find(&videos)
	log.Println(videos)
	return videos

}
