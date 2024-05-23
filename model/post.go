package model

type Post struct {
	ID             int        `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	Title          string     `json:"title" gorm:"not null; size:255"`
	Content        string     `json:"content" gorm:"not null; size:4000"`
	Comments       []*Comment `json:"comments,omitempty" gorm:"foreignkey:PostID"`
	CommentsLocked bool       `json:"commentsLocked"`
}
