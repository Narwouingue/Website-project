package router

import (
	"log"
	"net/http"

	"package/web/app/creatorPage"
	"package/web/app/creatorPrivatePage"
	"package/web/app/home"
	"package/web/app/logoff"
	"package/web/app/privateAccess"
	"package/web/app/publicAccess"
	categorySearch "package/web/app/searchByCategory"

	"github.com/gin-gonic/gin"
)

func Routes() *gin.Engine {
	rtr := gin.Default()

	rtr.MaxMultipartMemory = 50000 << 20
	var id, userName, category string

	//GET
	rtr.GET("/video/"+id, publicAccess.Handler)                           // viewing public video
	rtr.GET("/private/video/"+id, privateAccess.Handler)                  // viewing private video
	rtr.GET("/user/"+userName, creatorPage.Handler)                       // get the page of a crator
	rtr.GET("/user/"+userName+"/membersOnly", creatorPrivatePage.Handler) // get the private videos of a creator
	rtr.GET("/home", home.Handler)                                        // homepage
	rtr.GET("/logoff", logoff.Handler)                                    // log the user off
	rtr.GET("/category/"+category, categorySearch.Handler)

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
	return rtr
}