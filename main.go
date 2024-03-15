package main

import (
	"fmt"
	"log"
	"net/http"
	"package/platform/db"
	"package/platform/structs"
	"regexp"
	"time"

	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gopkg.in/gomail.v2"

	"os"

	"github.com/livekit/protocol/auth"

	"package/web/app/CreatorPrivatePage"
	"package/web/app/creatorPage"
	"package/web/app/home"
	"package/web/app/logoff"
	"package/web/app/privateAccess"
	"package/web/app/publicAccess"
	categorySearch "package/web/app/searchByCategory"
	"package/web/app/user"

	"github.com/joho/godotenv"
)

//Gestion des tokens ici

// Algo des abonnements ici

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Failed to load environment variables")
	}
	rtr := gin.Default()
	db.ConnectToDatabase()
	rtr.MaxMultipartMemory = 50000 << 20
	var id, creator, category string

	rtr.LoadHTMLGlob("web/template/*")

	// POSTPostVideos)
	rtr.POST("/signup", SignUp)
	rtr.POST("/signin", SignIn)
	rtr.POST("/resetPasswword", ResetPasswword)
	rtr.POST("/deleteUser", DeleteUser)
	rtr.POST("/deleteVideo", DeleteVideo)
	rtr.POST("/like", LikeVideo)
	rtr.POST("/tip", SendToken)
	rtr.POST("/modifyVideo", ModifyVideo)

	rtr.POST("/follow", Follow)
	rtr.POST("/unfollow", Unfollow)
	rtr.POST("/subscribe", Subscribe)
	rtr.POST("/unsubscribe", Unsubscribe)

	go TrendingAlgorithm()

	//GETS
	rtr.GET("/video/"+id, publicAccess.Handler)                         // viewing public video
	rtr.GET("/video/private/"+id, privateAccess.Handler)                // viewing private video
	rtr.GET("/creator/"+creator, creatorPage.Handler)                   // get the page of a crator
	rtr.GET("/creator/"+creator+"/private", CreatorPrivatePage.Handler) // get the private videos of a creator
	rtr.GET("/home", home.Handler)                                      // homepage
	rtr.GET("/logoff", logoff.Handler)                                  // log the user off
	rtr.GET("/category/"+category, categorySearch.Handler)              // search by category
	rtr.GET("/user", user.Handler)
	rtr.GET("/settings")

	rtr.GET("/getToken", func(c *gin.Context) {
		token, err := getJoinToken("my-room", "identity")
		if err != nil {
			log.Fatal("can't connect to the room")
		}
		c.JSON(http.StatusAccepted, gin.H{"token": token})
	})

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}

}
func getJoinToken(room, identity string) (string, error) {
	at := auth.NewAccessToken(os.Getenv("LIVEKIT_API_KEY"), os.Getenv("LIVEKIT_API_SECRET"))
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(time.Hour)

	log.Print(os.Getenv("LIVEKIT_API_KEY"))
	log.Print(os.Getenv("LIVEKIT_API_SECRET"))

	return at.ToJWT()
}

func Follow(c *gin.Context) {
	// check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not connected"})
		return
	}

	//Get the creator
	creatorName := c.GetString("creator")
	var creator structs.Creator
	err := db.Db.Table("creators").Where("username = ?", creatorName).First(&creator).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Creator doesn't exist"})
	}

	// Add the creator to the user's followings
	user.Followings = append(user.Followings, creator)
	creator.Followers++
	db.Db.Save("users")
	db.Db.Save("creators")

}

func Unfollow(c *gin.Context) {
	// check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not connected"})
		return
	}

	//Get the creator
	creatorName := c.GetString("creator")
	var creator structs.Creator
	err := db.Db.Table("creators").Where("username = ?", creatorName).First(&creator).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Creator doesn't exist"})
	}

	// Create the new following slice
	var result []structs.Creator
	for _, v := range user.Followings {
		if v.CreatorName != creator.CreatorName {
			result = append(result, v)
		}
	}

	// Remove the creator from the user's followings
	user.Followings = result
	creator.Followers--
	db.Db.Save("users")
	db.Db.Save("creators")
}

func Subscribe(c *gin.Context) {
	// check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not connected"})
		return
	}

	//Get the creator
	creatorName := c.GetString("creator")
	var creator structs.Creator
	db.Db.Table("creators").Where("username = ?", creatorName).First(&creator)

	// Add the creator to the user's subscribings
	user.Followings = append(user.Subscribings, creator)
	creator.Followers++
	db.Db.Save("users")
	db.Db.Save("creators")
}

func Unsubscribe(c *gin.Context) {
	// check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not connected"})
		return
	}
	//Get the creator
	creatorName := c.GetString("creator")
	var creator structs.Creator
	db.Db.Table("users").Where("username = ?", creatorName).First(&creator)

	// Create the new subscribing slice
	var result []structs.Creator
	for _, v := range user.Subscribings {
		if v.CreatorName != creator.CreatorName {
			result = append(result, v)
		}
	}

	// Remove the creator from the user's subscribings
	user.Subscribings = result
	creator.Followers--
	db.Db.Save("users")
	db.Db.Save("creators")
}

// TOKENS //
func SendToken(c *gin.Context) {
	// Check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		return
	}

	// Check if the user has enough tokens
	number := c.GetInt("number")
	if user.Tokens < number {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "not enough tokens available"})
		return
	}

	// check if creator exists
	creatorName := c.GetString("creator")
	var creator structs.Creator
	if err := db.Db.Table("creators").Where("creatorname = ?", creatorName).First(&creator).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "creator does not exist"})
		return
	}

	// Perform the transfer
	user.Tokens -= number
	userCreator := creator.User
	userCreator.Tokens += number

	// Send the confirmation and save the database
	c.JSON(http.StatusAccepted, gin.H{"message": "Tokens sent"})
	db.Db.Save("users")
	db.Db.Save("creators")
}

// VIDEOS //

func CreatorPrivateVideos(c *gin.Context) {

	// Check if creator exists
	creator := c.Param("creator")
	if err := db.Db.Table("creators").Where("creatorname = ?", creator).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "creator does not exist"})
		return
	}

	// check if user is subscribed
	connectedUser, message := GetConnectedUser(c)
	var count int
	for _, v := range connectedUser.Subscribings {
		if v.CreatorName == creator {
			break
		}
		if len(connectedUser.Subscribings) == count {
			c.JSON(http.StatusNotAcceptable, gin.H{"error": "You're not subscribed to this creator"})
			return
		}
		count++
	}

	//Fetch filepaths of private videos
	videos := getAllVideosFromACreator(creator)
	var privateVideos []string
	for _, v := range videos {
		if !v.IsPublic {
			privateVideos = append(privateVideos, v.FilePath)
		}
	}

	//Return private videos
	c.JSON(http.StatusAccepted, gin.H{"videos": privateVideos, "message": message})
}

func PrivateAccess(c *gin.Context) {

}

func PostVideos(c *gin.Context) {
	// Generate and ID
	newID := uuid.New().String()
	isIdUsed := db.Db.Table("videos").Where("id = ?", newID).First(&structs.Video{})
	for isIdUsed.Error == nil {
		newID = uuid.New().String()
		isIdUsed = db.Db.First("id = ?", newID)
		log.Println("boucle")

	}

	// single file
	file, _ := c.FormFile("file")
	log.Println(file.Filename)
	file.Filename = newID

	// Upload the file to specific dst.
	c.SaveUploadedFile(file, "./videos")
	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	title := c.GetString("Title")

	// check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		c.Redirect(http.StatusSeeOther, "/signin")
		return
	}

	// check if user is a creator
	var creator structs.Creator
	if db.Db.Where("user = ?", user).First(&creator).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are not a creator"})
		return
	}

	// declare variables
	currentTime := time.Now()
	date := currentTime.Format("02-01-2006")
	artist := user.FullName
	isPublic := c.GetBool("public")
	category := c.GetString("Category")

	//create the reference in the database
	db.Db.Table("videos").Create(&structs.Video{
		ID:       file.Filename,
		Title:    title,
		Owner:    user,
		Artist:   artist,
		Views:    0,
		Category: category,
		Date:     date,
		FilePath: "/videos/" + category + "/",
		IsPublic: isPublic})
}

func DeleteVideo(c *gin.Context) {
	// check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "user not connected"})
		return
	}

	// check if the user is the owner of the video
	var video structs.Video
	db.Db.Table("videos").Where("id = ?", c.GetString("id")).First(&video)
	if &user == &video.Owner || user.UserName == "root" {
		// delete the video
		db.Db.Table("videos").Where("id = ?", c.GetString("id")).Delete(&video)
	} else {
		// return an error
		c.JSON(http.StatusBadGateway, gin.H{"error": "You can't perform that action"})
		return
	}
}
func ModifyVideo(c *gin.Context) {
	// check if user is connected
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		c.Redirect(http.StatusSeeOther, "/signin")
		return
	}

	// check if the video exists
	id := c.GetString("id")
	var video structs.Video
	if db.Db.Table("id").Where("id = ?", id).First(&video).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video doesn't exist"})
		return
	}

	// check if the user is the owner of the video
	var creator structs.Creator
	db.Db.Table("creators").Where("user = ?", user).First(&creator)
	if &video.Owner != &creator.User {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed", "message": "You are not allowed to perform this action"})
		return
	}

	// get the new informations
	title := c.GetString("title")
	isPublic := c.GetBool("public")
	category := c.GetString("Category")

	//update the reference in the database
	video.Title = title
	video.Category = category
	video.FilePath = "/videos/" + category + "/"
	video.IsPublic = isPublic
	db.Db.Save(&video)

}

func LikeVideo(c *gin.Context) {
	// check if user is connected
	connectedUser, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user not connected"})
		return
	}

	// Get the video
	var video structs.Video
	id := c.GetString("id")
	db.Db.Table("videos").Where("id = ?", id).First(&video)

	// Remove the like
	if isLiked(c, connectedUser, id) {
		video.Likes--
		var newLikes []structs.Video
		for i := range connectedUser.Likes {
			if connectedUser.Likes[i].ID != video.ID {
				newLikes = append(newLikes, connectedUser.Likes[i])
			}
		}
		connectedUser.Likes = newLikes
		c.JSON(http.StatusOK, gin.H{"message": "like removed"})
	} else { // Add the like
		video.Likes++
		connectedUser.Likes = append(connectedUser.Likes, video)
		c.JSON(http.StatusOK, gin.H{"message": "like added"})
	}

	// Save the changes
	db.Db.Save(&connectedUser)
	db.Db.Save(&video)
	log.Println(video)

}

func isLiked(c *gin.Context, user structs.User, id string) bool {
	var video structs.Video
	db.Db.Table("videos").Where("id = ?", id).First(&video)
	for i := range user.Likes {
		if user.Likes[i].ID == video.ID {
			return true
		}
	}
	return false

}

func TrendingAlgorithm() {
	const (
		viewsWeight    = 0.3
		likesWeight    = 0.6
		commentsWeight = 0.1
	)
	cutoffDate := time.Now().AddDate(0, -2, 0)
	var videos []structs.Video
	var filteredVideos []structs.Video
	db.Db.Table("videos").Order("views desc").Find(&videos)
	for _, video := range videos {
		videoDate, err := parseDate(video.Date)
		if err != nil {
			log.Printf(video.ID, err)
			// Handle the error as needed
			continue
		}

		if videoDate.After(cutoffDate) {
			// Include the video in the filtered list
			filteredVideos = append(filteredVideos, video)
		}
	}

	for i := range filteredVideos {
		content := &filteredVideos[i]
		contentScore := viewsWeight*float64(content.Views) + likesWeight*float64(content.Likes) // + commentsWeight*float64(len(content.Comments))
		content.Score = contentScore
	}

	time.AfterFunc(time.Hour, TrendingAlgorithm)

}

func getAllVideosFromACreator(username string) []structs.Video {
	var videos []structs.Video
	db.Db.Table("videos").Where("owner = ?", username).Find(&videos)
	log.Println(videos)
	return videos

}

func SearchByCategory(c *gin.Context) {
	category := c.Param("category")
	var videos []structs.Video
	db.Db.Table("videos").Where("category = ?", category).Order("score desc").Find(&videos)

	log.Println(videos)

}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02-01-2006", dateStr)
}

/*------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------*/

// USERS //

func mdHashing(input string) string {
	byteInput := []byte(input)
	md5Hash := md5.Sum(byteInput)
	return hex.EncodeToString(md5Hash[:]) // by referring to it as a string
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

func DeleteUser(c *gin.Context) {
	id := c.GetString("id")
	token, err := c.Cookie("connection")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		return
	}
	var user structs.User
	err = db.Db.Table("users").Where("(email = ? OR username = ?) AND password = ?", id, id, mdHashing(c.GetString("password"))).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if token != user.AccessToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
		return
	}

	// Delete the user from the database
	db.Db.Table("users").Delete(&user)
	c.JSON(http.StatusAccepted, gin.H{"message": "user successfully deleted"})

}

func ChangeCreator(c *gin.Context) {
	user, message := GetConnectedUser(c)
	if message == "error" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		return
	}
	creatorName := c.GetString("creatorname")

	db.Db.Create(structs.Creator{
		CreatorName: creatorName,
		Followers:   0,
		Subscribers: 0,
		IsLive:      false,
	})

	db.Db.Save(&user)
	c.Redirect(http.StatusSeeOther, "/user")
}

func SignUp(c *gin.Context) {
	userName := c.GetString("username")
	usernamePattern := "^[a-zA-Z0-9_]+$"
	regex := regexp.MustCompile(usernamePattern)
	if !regex.MatchString(userName) && len(userName) <= 20 {
		c.JSON(http.StatusForbidden, gin.H{"error": "username not allowed", "message": "Username can only contains letters, numbers, underscores and should not exceed 15 characters"})
		return
	}

	if db.Db.Table("users").Where("username = ?", userName).First(&structs.User{}).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrongUsername", "message": "Chosen username is already in use"})
		return
	}

	email := c.GetString("email")
	if db.Db.Table("users").Where("username = ?", email).First(&structs.User{}).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrongEmail", "message": "Chosen email is already in use"})
		return
	}

	fullName := c.GetString("fullname")
	password := mdHashing(c.GetString("password"))

	randomBytes := make([]byte, 50)
	_, err := rand.Read(randomBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't create a token"})
	}
	token := base64.StdEncoding.EncodeToString(randomBytes)
	c.SetCookie("connection", token, 2628000, "/", "http:localhost", true, true)

	db.Db.Table("users").Create(&structs.User{UserName: userName, FullName: fullName, Email: email, Password: password, AccessToken: token})
	c.Redirect(http.StatusSeeOther, "/home")
}

func SignIn(c *gin.Context) {
	var user structs.User
	id := c.GetString("id")
	password := mdHashing(c.GetString("password"))
	if db.Db.Table("users").Where("(username = ? OR email = ?) AND password = ?", id, id, password).First(&user).Error == nil {

		token := user.AccessToken
		c.SetCookie("connection", token, 2628000, "/", "http:localhost", true, true)
	}
	c.Redirect(http.StatusSeeOther, "/home")
}

func ResetPasswword(c *gin.Context) {
	mail := c.GetString("mail")
	rtr := gin.Default()

	randomBytes := make([]byte, 50)
	_, err := rand.Read(randomBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't create a token to verify identity"})
	}
	token := base64.StdEncoding.EncodeToString(randomBytes)

	m := gomail.NewMessage()
	path := "/resetPassword/" + token
	url := "http://localhost:3000" + path

	if db.Db.Table("users").Where("mail = ?", mail).First(&structs.User{}).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email doesn't exist"})
	}

	m.SetHeader("From", "gnarwingg@gmail.com")

	m.SetHeader("To", mail)

	//m.SetAddressHeader("Cc", "oliver.doe@example.com", "Oliver")

	m.SetHeader("Password reset")

	m.SetBody("text/html", "Here's your password changing link : "+url)

	//m.Attach("lolcat.jpg")

	d := gomail.NewDialer("smtp.gmail.com", 587, "gnarwingg@gmail.com", "qmtzmfthrqlnvqlx")

	// Send the email to Kate, Noah and Oliver.

	if err := d.DialAndSend(m); err != nil {

		panic(err)
	}

	rtr.GET(path, func(c *gin.Context) {

		rtr.POST("/changePassword", func(c *gin.Context) {
			newpassword := c.GetString("newspassword")
			var user structs.User
			db.Db.Table("users").Where("email = ?", mail).First(&user)
			user.Password = newpassword
			db.Db.Save(&user)
			c.JSON(http.StatusOK, gin.H{"message": "Password successfully changed"})
			c.Redirect(http.StatusSeeOther, "/home")
		})

	})
}
