package router

import (
	"log"
	"net/http"
	handler "package/handlers"

	"github.com/gin-gonic/gin"
)

func Routes() {
	rtr := gin.Default()

	rtr.MaxMultipartMemory = 50000 << 20
	var id, userName, category string
	// POST
	rtr.POST("/upload", handler.PostVideos)
	rtr.POST("/signup", handler.SignUp)
	rtr.POST("/signin", handler.SignIn)
	rtr.POST("/resetPasswword", handler.ResetPasswword)
	rtr.POST("/deleteUser", handler.DeleteUser)
	rtr.POST("/deleteVideo", handler.DeleteVideo)
	rtr.POST("/like", handler.LikeVideo)
	rtr.POST("/dislike", handler.DislikeVideo)
	rtr.POST("/tip", handler.SendToken)
	rtr.POST("/modifyVideo", handler.ModifyVideo)

	//GET
	rtr.GET("/video/"+id, handler.PublicAccess)                             // viewing public video
	rtr.GET("/private/video/"+id, handler.PrivateAccess)                    // viewing private video
	rtr.GET("/user/"+userName, handler.CreatorPage)                         // get the page of a crator
	rtr.GET("/user/"+userName+"/membersOnly", handler.CreatorPrivateVideos) // get the private videos of a creator
	rtr.GET("/home", handler.Home)                                          // homepage
	rtr.GET("/logoff", handler.Logoff)                                      // log the user off
	rtr.GET("/category/"+category, handler.SearchByCategory)

	go handler.TrendingAlgorithm()

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
