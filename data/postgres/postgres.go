package postgres

import (
	"errors"
	"fmt"
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
		PostID:  postID,
		Content: content,
	}
	err := tx.Create(&newComment).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("can't insert comment: " + err.Error())
	}
	if post.HasComments == false {
		post.HasComments = true
		tx.Save(&post)
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

func (d Db) commentProcess(comments []model.Comment, limit int) []*model.Comment {
	var res []*model.Comment
	commentsMap := make(map[uint]*model.Comment)
	for _, comment := range comments {
		if comment.ParentID == nil {
			if len(res) < limit {
				res = append(res, &comment)
			}
		} else if value, ok := commentsMap[*comment.ParentID]; ok {
			value.Replies = append(value.Replies, &comment)
		}
		commentsMap[comment.ID] = &comment
	}
	return res
}

func (d Db) Post(id uint, limit *int) (*model.Post, error) {
	var post model.Post
	if err := d.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	if limit == nil || *limit <= 0 {
		limit = new(int)
		*limit = 10
	}
	var comments []model.Comment
	err := d.db.Where("post_id = ?", id).Order("id ASC").Find(&comments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &post, nil
		}
		return nil, err
	}
	post.Comments = d.commentProcess(comments, *limit)
	return &post, nil

}

func (d Db) Posts() ([]*model.Post, error) {
	posts := make([]*model.Post, 0)
	res := d.db.Model(&model.Post{}).Order("id ASC").Find(&posts)
	if res.Error != nil {
		return nil, res.Error
	}
	return posts, nil
}

func (d Db) Comments(postID uint, first *int, after *int) (*model.CommentConnection, error) {
	if first == nil || *first <= 0 {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = new(int)
		*after = 0
	} else if *after < 0 {
		return nil, errors.New(fmt.Sprintf("comment with id %d not found", *after))
	} else {
		var checkComment model.Comment
		err := d.db.Where("id = ?", *after).First(&checkComment).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New(fmt.Sprintf("comment with id %d not found", *after))
			}
			return nil, err
		}
	}

	var comments []model.Comment
	err := d.db.Where("post_id = ? and id > ?", postID, *after).Order("id").Limit(*first + 1).Find(&comments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model.CommentConnection{
		Comments: d.commentProcess(comments, *first),
		PageInfo: &model.PageInfo{
			HasNextPage: len(comments) > *first,
		},
	}, nil

}

func New(cfg string) *Db {
	db, err := gorm.Open(postgres.Open(cfg), &gorm.Config{})
	if err != nil {
		panic("couldn't connect to database: " + err.Error())
	}
	err = db.AutoMigrate(&model.Post{}, &model.Comment{})
	if err != nil {
		panic("failed to migrate tables: " + err.Error())
	}
	db = db.Debug()
	return &Db{db: db}
}

func (d Db) Stop() error {
	val, err := d.db.DB()
	if err != nil {
		return errors.New("failed to get database error: " + err.Error())
	}
	return val.Close()
}

//func (d Db) Reply(obj *model.Comment) (*model.Comment, error) {
//	var res *model.Comment
//	err := d.db.Where("parent_id_i = ?", obj.ID).Order("id ASC").First(&res).Error
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return nil, nil
//		}
//		return nil, err
//	}
//
//	return res, nil
//}
