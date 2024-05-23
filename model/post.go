package model

type Post struct {
	ID             uint      `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	Title          string    `json:"title" gorm:"not null; size:255"`
	Content        string    `json:"content" gorm:"not null; size:4000"`
	HasComments    bool      `json:"hasComments" gorm:"not null; default:false"`
	CommentsLocked bool      `json:"commentsLocked" gorm:"not null; default:false"`
	Comments       []Comment `json:"replies" gorm:"-"`
}
