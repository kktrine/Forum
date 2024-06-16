```graphql
mutation CreatePost {
    createPost(title: "post", content: "post content", commentsLocked: true) {
        id
        title
        content
        commentsLocked
        hasComments
    }
}

```

```graphql
query GetPosts {
    posts {
        id
        title
        content
        commentsLocked
        hasComments
    }
}

```

```graphql
mutation LockPostComments {
    lockComments(postID: 1) {
        id
        title
        content
        commentsLocked
        hasComments
    }
}

```
```graphql
query GetOnePost {
    post(id: 3) {
        id
        title
        content
        hasComments
        commentsLocked
        comments {
            id
            content
            replies {
                id
                content
                replies {
                    id
                    content
                    replies {
                        id
                        content
                    }
                }
            }
        }
    }
}

```
```graphql
mutation CreateComment {
    createComment(postID: 3, content: "comment") {
        id
        postId
        parentId
        content
    }
}

```
```graphql
query GetComments {
    comments(postID: 3, first: -3, afterCursor: "26") {
        comments {
            id
            content
            replies {
                id
                content
            }
        }
        pageInfo {
            hasNextPage
            endCursor
        }
    }
}

```
```graphql
subscription Subscribe {
    newComment(postID: "2") {
        id
        content
    }
}


```






