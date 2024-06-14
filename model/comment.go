package model

type Comment struct {
	ID       uint       `json:"id" gorm:"primary_key;auto_increment"`
	PostID   uint       `json:"postId" gorm:"foreignkey:PostID"`
	ParentID *uint      `json:"parentIdI,omitempty" gorm:"index"`
	Content  string     `json:"content" gorm:"type:text;not null;size:2000"`
	Replies  []*Comment `json:"replies,omitempty" gorm:"-"`
}
