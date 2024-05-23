package model

type Comment struct {
	ID       uint     `json:"id" gorm:"primary_key;auto_increment"`
	PostID   uint     `json:"postId" gorm:"foreignkey:PostID"`
	ParentID *uint    `json:"parentId,omitempty"`
	Content  string   `json:"content" gorm:"type:text;not null;size:2000"`
	Reply    *Comment `json:"replies,omitempty" gorm:"foreignkey:ParentID"`
}
