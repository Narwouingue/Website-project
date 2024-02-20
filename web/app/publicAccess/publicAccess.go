package publicAccess

import (
	"net/http"
	"package/db"
	"package/structs"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	id := c.Param("id")
	var video structs.Video
	filepath := db.Db.Table("videos").Where("id = ?", id).First(&video)
	if filepath.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "404 not found"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"filepath": filepath})
	video.Views++
	db.Db.Save("videos")

}
