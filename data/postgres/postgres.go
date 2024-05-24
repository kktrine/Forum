package postgres

import (
	"errors"
	"forum/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Db struct {
	db *gorm.DB
}

func (d Db) CreatePost(title string, content string, commentsLocked *bool) (*model.Post, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	newPost := &model.Post{
		Title:       title,
		Content:     content,
		HasComments: false,
		Comments:    nil,
	}
	if commentsLocked != nil {
		newPost.CommentsLocked = *commentsLocked
	}
	if err := tx.Create(&newPost).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("can't insert data: " + err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("can't commit transaction: " + err.Error())
	}
	return newPost, nil
}

func (d Db) CreateComment(postID uint, parentIDI *uint, _ *string, content string) (*model.Comment, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	var post model.Post
	if err := d.db.First(&post, postID).Error; err != nil {
		return nil, errors.New("can't find post: " + err.Error())
	}
	if parentIDI != nil {
		var comment model.Comment
		err := d.db.First(&comment, parentIDI).Error
		if err != nil {
			return nil, errors.New("can't find comments: " + err.Error())
		}
	}
	if post.CommentsLocked {
		return nil, errors.New("can't comment this post")
	}
	newComment := model.Comment{
		PostID:    postID,
		ParentIDI: parentIDI,
		Content:   content,
	}
	err := tx.Create(&newComment).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("can't insert comment: " + err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("can't commit transaction: " + err.Error())
	}
	return &newComment, nil
}

func (d Db) LockComments(postID uint) (*model.Post, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	var post model.Post
	if err := d.db.First(&post, postID).Error; err != nil {
		return nil, err
	}
	post.CommentsLocked = true
	if err := d.db.Save(&post).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("can't commit transaction: " + err.Error())
	}
	return &post, nil
}

func (d Db) Post(id uint) (*model.Post, error) {
	var post model.Post
	if err := d.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	res := d.db.Where("Post_ID = ? and Parent_id_i is null", id).Find(&post.Comments)
	if res.Error != nil {
		return nil, errors.New("something went wrong: " + res.Error.Error())
	}
	return &post, nil
}

func (d Db) Posts() ([]*model.Post, error) {
	posts := make([]*model.Post, 0)
	res := d.db.Order("id").Find(&posts)
	if res.Error != nil {
		return nil, res.Error
	}
	return posts, nil
}

func (d Db) Comments(id *uint, first *int, after *string) (*model.CommentConnection, error) {
	//TODO implement me
	panic("implement me")
}

func New(cfg string) *Db {
	db, err := gorm.Open(postgres.Open(cfg), &gorm.Config{})
	if err != nil {
		panic("coudn't connect to database: " + err.Error())
	}
	err = db.AutoMigrate(&model.Post{}, &model.Comment{})
	if err != nil {
		panic("failed to migrate tables: " + err.Error())
	}
	return &Db{db: db}
}

func (d Db) Stop() error {
	val, err := d.db.DB()
	if err != nil {
		return errors.New("failed to get database error: " + err.Error())
	}
	return val.Close()
}
