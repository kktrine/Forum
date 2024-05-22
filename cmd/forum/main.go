package main

import "forum/internal/forumSQL"

func main() {
	dbConf := "host=localhost user=postgres password=postgres dbname=forum port=5432 sslmode=disable"
	gb := forumSQL.NewDataBase(dbConf)
	_ = gb
}
