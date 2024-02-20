package handler

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"time"

	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dsn = "root:2580@tcp(localhost:8080)/data?charset=utf8mb4&parseTime=True&loc=Local"
var db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

func main() {
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&video{}, &user{})

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

}

// TOKENS //
func SendToken(c *gin.Context) {
	creatorName := c.GetString("creator")
	number := c.GetInt("number")
	token, err := c.Cookie("connection")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		return
	}
	var user, creator user
	db.Table("users").Where("token = ?", token).First(&user)
	if user.Tokens < number {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "not enough tokens available"})
		return
	}

	db.Table("users").Where("username = ?", creatorName).First(&creator)
	if !creator.IsCreator {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "Bad sending", "message": "The user you are trying to send tokens to is not a creator"})
		return
	}

	user.Tokens -= number
	creator.Tokens += number
	db.Save("users")
	c.JSON(http.StatusAccepted, gin.H{"message": "Tokens sent"})
}

//Gestion des tokens ici

// Algo des abonnements ici

// VIDEOS //
type video struct {
	ID          string  `gorm:"column:id" json:"id"`
	Folder      string  `gorm:"folder" json:"folder"`
	Title       string  `gorm:"column:title" json:"title"`
	Artist      string  `gorm:"column:artist" json:"artist"`
	Owner       string  `gorm:"column:owner" json:"owner"`
	Date        string  `gorm:"column:date" json:"date"`
	Category    string  `gorm:"column:category" json:"category"`
	Views       int     `gorm:"column:views" json:"views"`
	Likes       int     `gorm:"column:likes" json:"likes"`
	Dislikes    int     `gorm:"column:dislikes" json:"dislikes"`
	Score       float64 `gorm:"column:score" json:"score"`
	Description string  `gorm:"column:description" json:"description"`
	FilePath    string  `gorm:"colum:filepath" json:"filepath"`
	IsPublic    bool    `gorm:"column:ispublic"`

	//Comments []string `gorm:"column:comments" json:"comments"`
}

func CreatorPrivateVideos(c *gin.Context) {
	userName := c.Param("userName")
	var user user

	err := db.Table("users").Where("username = ?", userName).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Creator not found"})
		return
	}
	if !user.IsCreator {
		c.JSON(http.StatusNotFound, gin.H{"error": "This user is not a creator"})
		return
	}

	connectedUser, message := GetConnectedUser(c)

	err = db.Table("users").Where("username = ?", userName).Where("subcribers = ?", connectedUser).Find(&user).Error
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "You're not subscribed to this creator"})
		return
	}
	videos := getAllVideosFromACreator(userName)
	var privateVideos []video
	for _, v := range videos {
		if !v.IsPublic {
			privateVideos = append(privateVideos, v)
		}
	}

	c.JSON(http.StatusAccepted, gin.H{"videos": privateVideos, "message": message})
}

func PublicAccess(c *gin.Context) {
	id := c.Param("id")
	var video video
	filepath := db.Table("videos").Where("id = ?", id).First(&video)
	if filepath.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "404 not found"})
		return
	} else if video.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to be there"})
	}
	c.JSON(http.StatusAccepted, gin.H{"filepath": filepath})
	watchVideo(id)

}

func PrivateAccess(c *gin.Context) {

}

func PostVideos(c *gin.Context) {
	newID := uuid.New().String()
	isIdUsed := db.Table("videos").Where("id = ?", newID).First(&video{})

	for isIdUsed.Error == nil {
		newID = uuid.New().String()
		isIdUsed = db.First("id = ?", newID)
		log.Println("boucle")

	}
	currentTime := time.Now()
	date := currentTime.Format("02-01-2006")

	// single file
	file, _ := c.FormFile("file")

	log.Println(file.Filename)
	file.Filename = newID
	// Upload the file to specific dst.
	c.SaveUploadedFile(file, "./videos")

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	title := c.GetString("Title")

	token, err := c.Cookie("connection")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		c.Redirect(http.StatusSeeOther, "/signin")
		return
	}

	var user user
	db.Table("users").Where("token = ?", token).First(&user)
	if !user.IsCreator {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not a creator"})
		return
	}
	owner := user.UserName
	artist := user.FullName
	isPublic := c.GetBool("public")
	category := c.GetString("Category")
	//sstring userId = User.Claims.FirstOrDefault(c => c.Type == ClaimTypes.NameIdentifier).Value;

	//create the reference in the database
	db.Table("videos").Create(&video{ID: file.Filename, Title: title, Owner: owner, Artist: artist, Views: 0, Category: category, Date: date, FilePath: "/videos/" + category + "/", IsPublic: isPublic})
}

func DeleteVideo(c *gin.Context) {
	token, err := c.Cookie("connection")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "user not connected"})
		return
	}
	var user user
	db.Table("users").Where("token = ?", token).First(&user)
	videoOwner := c.GetString("owner")

	if user.UserName == videoOwner || user.UserName == "root" {
		db.Table("videos").Where("id = ?", c.GetString("id")).Delete(&video{})
	} else if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "You can't perform that action"})
		return

	}

}
func ModifyVideo(c *gin.Context) {
	token, err := c.Cookie("connection")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		c.Redirect(http.StatusSeeOther, "/signin")
		return
	}

	var user user
	db.Table("users").Where("token = ?", token).First(&user)
	if !user.IsCreator {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not a creator"})
		return
	}
	owner := user.UserName

	title := c.GetString("title")
	videoOwner := c.GetString("owner")
	isPublic := c.GetBool("public")
	category := c.GetString("Category")
	if owner != videoOwner && user.UserName != "root" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed", "message": "You are not allowed to perform this action"})
		return
	}
	id := c.GetString("id")
	var video video
	if db.Table("id").Where("id = ?", id).First(&video).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video doesn't exist"})
		return
	}

	//update the reference in the database
	video.Title = title
	video.Category = category
	video.FilePath = "/videos/" + category + "/"
	video.IsPublic = isPublic

	db.Save(&video)

}

func LikeVideo(c *gin.Context) {
	var video, isLiked video
	id := c.GetString("id")
	connectedUser, message := GetConnectedUser(c)
	db.Table("videos").Where("id = ?", id).First(&video)
	if message == "error" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user not connected"})
		return
	}

	err := db.Table("users").Where("username = ?", connectedUser.UserName).Where("likes = ?", id).First(&isLiked).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "video unliked"})
		video.Likes--
		db.Save(&video)
		return
	} else if db.Table("users").Where("username = ?", connectedUser.UserName).Where("dislikes = ?", id).First(&isLiked).Error != nil {
		video.Likes++
		video.Dislikes--
		db.Table("users").Where("username = ?", connectedUser.UserName).Where("dislikes = ?", id).Delete(&isLiked)
		db.Save(&video)
		return
	}
	video.Likes++
	db.Save(&video)

	log.Println(video)

}

func DislikeVideo(c *gin.Context) {
	var video, isLiked video
	id := c.GetString("id")
	connectedUser, message := GetConnectedUser(c)
	db.Table("videos").Where("id = ?", id).First(&video)
	if message == "error" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user not connected"})
		return
	}

	err := db.Table("users").Where("username = ?", connectedUser.UserName).Where("dislikes = ?", id).First(&isLiked).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "video unliked"})
		video.Dislikes--
		db.Save(&video)
		return
	} else if db.Table("users").Where("username = ?", connectedUser.UserName).Where("likes = ?", id).First(&isLiked).Error != nil {
		video.Likes--
		video.Dislikes++
		db.Table("users").Where("username = ?", connectedUser.UserName).Where("likes = ?", id).Delete(&isLiked)
		db.Save(&video)
		return
	}
	video.Dislikes++

	db.Save(&video)

	log.Println(video)

}

func TrendingAlgorithm() {
	const (
		viewsWeight    = 0.3
		likesWeight    = 0.6
		commentsWeight = 0.1
	)
	cutoffDate := time.Now().AddDate(0, -2, 0)
	var videos []video
	var filteredVideos []video
	db.Table("videos").Order("views desc").Find(&videos)
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

func Home(c *gin.Context) {
	trending := returnTrendingVideos()
	new := returnNewVideos()

	c.JSON(http.StatusOK, gin.H{"trending": trending, "new": new})
}

// Non rooter functions
func returnTrendingVideos() []video {
	var videos []video
	db.Table("videos").Order("score desc").Find(&videos)
	log.Println(videos)
	return videos
}

func returnNewVideos() []video {
	var videos []video
	db.Table("videos").Find(&videos)

	// Custom sorting function to order videos by date in descending order
	sort.Slice(videos, func(i, j int) bool {
		// Convert date strings to time.Time for comparison
		date1, err := parseDate(videos[i].Date)
		date2, err := parseDate(videos[j].Date)
		// Order by date in descending order
		if err != nil {
			log.Println("noooooo there's an error motherfucker")
		}
		return date1.After(date2)
	})

	log.Println(videos)
	return videos
}

func getAllVideosFromACreator(username string) []video {
	var videos []video
	db.Table("videos").Where("owner = ?", username).Find(&videos)
	log.Println(videos)
	return videos

}

func SearchByCategory(c *gin.Context) {
	category := c.Param("category")
	var videos []video
	db.Table("videos").Where("category = ?", category).Order("score desc").Find(&videos)

	log.Println(videos)
	return
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02-01-2006", dateStr)
}

func watchVideo(id string) {
	var video video
	db.Table("videos").Where("id = ?", id).First(&video)
	video.Views++
	db.Save(&video)
	log.Println(video)

}

/*------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------*/

// USERS //
type user struct {
	UserName      string   `gorm:"column:username"`
	FullName      string   `gorm:"column:fullname"`
	Email         string   `gorm:"column:email"`
	password      string   `gorm:"column:password"`
	Likes         []video  `gorm:"column:likes"`
	WatchedVideos []video  `gorm:"column:watchedvideos"`
	roles         []string `gorm:"column:roles"`
	AccessToken   string   `gorm:"column:token"`
	IsCreator     bool     `gorm:"colum:isCreator"`
	Subsribers    []user   `gorm:"column:subs"`
	Tokens        int      `gorm:"column:tokens"`
}

func mdHashing(input string) string {
	byteInput := []byte(input)
	md5Hash := md5.Sum(byteInput)
	return hex.EncodeToString(md5Hash[:]) // by referring to it as a string
}

func GetConnectedUser(c *gin.Context) (returnedUser user, message string) {
	token, err := c.Cookie("connection")

	if err != nil {
		var user user
		return user, "error"
	}
	var user user
	db.Table("users").Where("token = ?", token).First(&user)
	return user, "success"

}

func CreatorPage(c *gin.Context) {
	userName := c.Param("userName")
	var user user
	err := db.Table("users").Where("username = ?", userName).First(&user).Error
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

func DeleteUser(c *gin.Context) {
	id := c.GetString("id")
	token, err := c.Cookie("connection")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
		return
	}
	var user user
	err = db.Table("users").Where("(email = ? OR username = ?) AND password = ?", id, id, mdHashing(c.GetString("password"))).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if token != user.AccessToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
		return
	}

	// Delete the user from the database
	db.Table("users").Delete(&user)
	c.JSON(http.StatusAccepted, gin.H{"message": "user successfully deleted"})

}

func ChangeCreator(c *gin.Context) {
	token, err := c.Cookie("connection")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not connected"})
	}
	var user user
	db.Table("users").Where("token = ?", token).First(&user)
	user.IsCreator = !user.IsCreator
	db.Save(&user)
	c.Redirect(http.StatusSeeOther, "/user")
}

func getVideos() {
	//c.JSON(http.StatusOK, videos)
}

func SignUp(c *gin.Context) {
	userName := c.GetString("username")
	usernamePattern := "^[a-zA-Z0-9_]+$"
	regex := regexp.MustCompile(usernamePattern)
	if !regex.MatchString(userName) && len(userName) <= 20 {
		c.JSON(http.StatusForbidden, gin.H{"error": "username not allowed", "message": "Username can only contains letters, numbers, underscores and should not exceed 15 characters"})
		return
	}

	if db.Table("users").Where("username = ?", userName).First(&user{}).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrongUsername", "message": "Chosen username is already in use"})
		return
	}

	email := c.GetString("email")
	if db.Table("users").Where("username = ?", email).First(&user{}).Error != nil {
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

	db.Table("users").Create(&user{UserName: userName, FullName: fullName, Email: email, password: password, AccessToken: token})
	c.Redirect(http.StatusSeeOther, "/home")
}

func SignIn(c *gin.Context) {
	var user user
	id := c.GetString("id")
	password := mdHashing(c.GetString("password"))
	if db.Table("users").Where("(username = ? OR email = ?) AND password = ?", id, id, password).First(&user).Error == nil {

		token := user.AccessToken
		c.SetCookie("connection", token, 2628000, "/", "http:localhost", true, true)
	}
	c.Redirect(http.StatusSeeOther, "/home")
}

func Logoff(c *gin.Context) {
	token, err := c.Cookie("connection")
	if err != nil {
		c.SetCookie("connection", token, -500, "/", "http:localhost", true, true)
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

	if db.Table("users").Where("mail = ?", mail).First(&user{}).Error != nil {
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
			var user user
			db.Table("users").Where("email = ?", mail).First(&user)
			user.password = newpassword
			db.Save(&user)
			c.JSON(http.StatusOK, gin.H{"message": "Password successfully changed"})
			c.Redirect(http.StatusSeeOther, "/home")
		})

	})
}
