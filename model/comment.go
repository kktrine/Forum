package model

type Comment struct {
	ID       int        `json:"id" gorm:"primary_key;auto_increment"`
	PostID   int        `json:"postId" gorm:"foreignkey:PostID"`
	ParentID *int       `json:"parentId,omitempty"`
	Content  string     `json:"content" gorm:"type:text;not null;size:2000"`
	Replies  []*Comment `json:"replies,omitempty" gorm:"foreignkey:ParentID"`
}
