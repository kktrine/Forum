package forumMemory

import "github.com/graphql-go/graphql"

type Post struct {
	ID             string
	Title          string
	Content        string
	Comments       []*Comment
	CommentsLocked bool
}

type Comment struct {
	ID       string
	PostID   string
	ParentID *string
	Content  string
	Replies  []*Comment
}

func createSchema() graphql.Schema {
	commentType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Comment",
		Fields: graphql.Fields{
			"id":      &graphql.Field{Type: graphql.ID},
			"content": &graphql.Field{Type: graphql.String},
			"replies": &graphql.Field{
				Type: graphql.NewList(commentType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					comment := p.Source.(*Comment)
					return getReplies(comment.ID), nil
				},
			},
		},
	})

	postType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Post",
		Fields: graphql.Fields{
			"id":      &graphql.Field{Type: graphql.ID},
			"title":   &graphql.Field{Type: graphql.String},
			"content": &graphql.Field{Type: graphql.String},
			"comments": &graphql.Field{
				Type: graphql.NewList(commentType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					post := p.Source.(*Post)
					return getCommentsByPostID(post.ID), nil
				},
			},
			"commentsLocked": &graphql.Field{Type: graphql.Boolean},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"posts": &graphql.Field{
				Type: graphql.NewList(postType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return posts, nil
				},
			},
			"post": &graphql.Field{
				Type: postType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(string)
					return getPostByID(id), nil
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createPost": &graphql.Field{
				Type: postType,
				Args: graphql.FieldConfigArgument{
					"title":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"content": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					title := p.Args["title"].(string)
					content := p.Args["content"].(string)
					newPost := &Post{
						ID:      uuid.New().String(),
						Title:   title,
						Content: content,
					}
					posts = append(posts, newPost)
					return newPost, nil
				},
			},
			"createComment": &graphql.Field{
				Type: commentType,
				Args: graphql.FieldConfigArgument{
					"postID":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"parentID": &graphql.ArgumentConfig{Type: graphql.ID},
					"content":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					postID := p.Args["postID"].(string)
					parentID, _ := p.Args["parentID"].(string)
					content := p.Args["content"].(string)

					newComment := &Comment{
						ID:       uuid.New().String(),
						PostID:   postID,
						Content:  content,
						ParentID: &parentID,
					}

					if parentID == "" {
						newComment.ParentID = nil
					}

					comments = append(comments, newComment)
					return newComment, nil
				},
			},
			"lockComments": &graphql.Field{
				Type: postType,
				Args: graphql.FieldConfigArgument{
					"postID": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					postID := p.Args["postID"].(string)
					post := getPostByID(postID)
					if post != nil {
						post.CommentsLocked = true
					}
					return post, nil
				},
			},
		},
	})

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	return schema
}
