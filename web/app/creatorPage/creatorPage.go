package creatorPage

import (
	"log"
	"net/http"
	"package/platform/db"
	"package/platform/structs"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	userName := c.Param("userName")

	var creator structs.Creator
	if err := db.Db.Table("creators").Where("creatorname = ?", userName).First(&creator).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "creator does not exist"})
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
