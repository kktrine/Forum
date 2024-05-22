package forumSQL

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Post struct {
	ID              uint      `gorm:"primarykey; auto_increment:true"`
	Title           string    `gorm:"size:255; not null"`
	Content         string    `gorm:"size:2000; not null"`
	CommentsEnabled bool      `gorm:"default:true"`
	Comments        []Comment `gorm:"foreignkey:PostID"`
}

type Comment struct {
	ID       uint      `gorm:"primarykey; auto_increment:true"`
	PostID   uint      `gorm:"foreignKey:PostID"`
	ParentID *uint     `gorm:"foreignKey:ParentID"`
	Content  string    `gorm:"size:2000; not null"`
	Replies  []Comment `gorm:"foreignkey:ParentID"`
}
type Resolver struct {
	DB *gorm.DB
}

type QueryResolver struct{ *Resolver }
type MutationResolver struct{ *Resolver }

func NewDataBase(dbData string) *Resolver {
	db, err := gorm.Open(postgres.Open(dbData), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	err = db.AutoMigrate(&Post{}, &Comment{})
	if err != nil {
		panic("failed to migrate tables: " + err.Error())
	}
	return &Resolver{DB: db}
}

func (r *Resolver) Query() QueryResolver {
	return QueryResolver{r}
}

func (r *Resolver) Mutation() MutationResolver {
	return MutationResolver{r}
}

func (r *QueryResolver) Posts() []*Post {
	var posts []*Post
	r.DB.Preload("Comments").Find(&posts)
	return posts
}

func (r *QueryResolver) Post(id string) *Post {
	var post Post
	r.DB.Preload("Comments").First(&post, id)
	return &post
}

func (r *MutationResolver) CreatePost(title string, content string, commentsEnabled bool) (*Post, error) {
	post := &Post{Title: title, Content: content, CommentsEnabled: commentsEnabled}
	if err := r.DB.Create(post).Error; err != nil {
		return nil, err
	}
	return post, nil
}

func (r *MutationResolver) CreateComment(postID uint, parentID *uint, content string) (*Comment, error) {
	comment := &Comment{PostID: postID, ParentID: parentID, Content: content}
	if err := r.DB.Create(comment).Error; err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *MutationResolver) DisableComments(postID uint) (*Post, error) {
	var post Post
	err := r.DB.First(&post, postID).Error
	if err != nil {
		return nil, err
	}
	post.CommentsEnabled = false
	err = r.DB.Save(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}
