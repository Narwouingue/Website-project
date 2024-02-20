package home

import (
	"log"
	"net/http"
	"package/db"
	"package/structs"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	trending := returnTrendingVideos()
	new := returnNewVideos()

	c.JSON(http.StatusOK, gin.H{"trending": trending, "new": new})
}

func returnTrendingVideos() []structs.Video {
	var videos []structs.Video
	db.Db.Table("videos").Order("score desc").Find(&videos)
	log.Println(videos)
	return videos
}
func returnNewVideos() []structs.Video {
	var videos []structs.Video
	db.Db.Table("videos").Find(&videos)

	// Custom sorting function to order videos by date in descending order
	sort.Slice(videos, func(i, j int) bool {
		// Convert date strings to time.Time for comparison
		date1, err := parseDate(videos[i].Date)
		if err != nil {
			log.Println("error 1")
		}
		date2, err := parseDate(videos[j].Date)
		// Order by date in descending order
		if err != nil {
			log.Println("error 2")
		}
		return date1.After(date2)
	})

	log.Println(videos)
	return videos
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02-01-2006", dateStr)
}
