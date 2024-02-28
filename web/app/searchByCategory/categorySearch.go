package categorySearch

import (
	"log"
	"package/platform/db"
	"package/platform/structs"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	category := c.Param("category")
	var videos []structs.Video
	db.Db.Table("videos").Where("category = ?", category).Order("score desc").Find(&videos)

	log.Println(videos)

}
