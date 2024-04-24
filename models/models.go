package models

import (
	"github.com/watain666/ptt-alertor/models/article"
	"github.com/watain666/ptt-alertor/models/board"
	"github.com/watain666/ptt-alertor/models/user"
)

var User = func() *user.User {
	return user.NewUser(new(user.Redis))
}
var Article = func() *article.Article {
	return article.NewArticle(new(article.DynamoDB))
}
var Board = func() *board.Board {
	return board.NewBoard(new(board.DynamoDB), new(board.Redis))
}
