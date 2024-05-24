package model

type Comment struct {
	ID        uint     `json:"id" gorm:"primary_key;auto_increment"`
	PostID    uint     `json:"postId" gorm:"foreignkey:PostID"`
	ParentIDI *uint    `json:"parentIdI,omitempty" gorm:"index"`
	ParentIDS *uint    `json:"parentIdS,omitempty" gorm:"-"`
	Content   string   `json:"content" gorm:"type:text;not null;size:2000"`
	Reply     *Comment `json:"replies,omitempty" gorm:"foreignkey:ParentIDI"`
}
