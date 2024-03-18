package structs

type Video struct {
	ID          string  `gorm:"column:id" json:"id"`
	Folder      string  `gorm:"folder" json:"folder"`
	Title       string  `gorm:"column:title" json:"title"`
	Artist      string  `gorm:"column:artist" json:"artist"`
	Owner       Creator `gorm:"column:owner" json:"owner"`
	Date        string  `gorm:"column:date" json:"date"`
	Category    string  `gorm:"column:category" json:"category"`
	Views       int     `gorm:"column:views" json:"views"`
	Likes       int     `gorm:"column:likes" json:"likes"`
	Score       float64 `gorm:"column:score" json:"score"`
	Description string  `gorm:"column:description" json:"description"`
	FilePath    string  `gorm:"colum:filepath" json:"filepath"`
	IsPublic    bool    `gorm:"column:ispublic"`

	//Comments []string `gorm:"column:comments" json:"comments"`
}

type User struct {
	UserName      string    `gorm:"column:username"`
	FullName      string    `gorm:"column:fullname"`
	Email         string    `gorm:"column:email"`
	Password      string    `gorm:"column:password"`
	Likes         []Video   `gorm:"column:likes"`
	WatchedVideos []Video   `gorm:"column:watchedvideos"`
	AccessToken   string    `gorm:"column:token"`
	Followings    []Creator `gorm:"column:followings"`
	Subscribings  []Creator `gorm:"column:subscribings"`
	Tokens        int       `gorm:"column:tokens"`
}

type Creator struct {
	User        User   `gorm:"column:user"`
	CreatorName string `gorm:"column:creatorname"`
	Followers   int    `gorm:"column:followers"`
	Subscribers int    `gorm:"column:subscribers"`
	IsLive      bool   `gorm:"column:islive"`
}
