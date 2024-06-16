package main

import (
	"flag"
	"forum/internal/data"
	"forum/internal/data/memoryDB"
	"forum/internal/data/postgres"
	"forum/internal/graphQL"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	db := "host=localhost user=postgres password=postgres dbname=forum port=5432 sslmode=disable"
	storageFlag := flag.String("db", "sql", "sql or mem")
	flag.Parse()
	var storage data.Storage
	if *storageFlag == "sql" {
		storage = postgres.New(db)
	} else if *storageFlag == "mem" {
		storage = memoryDB.New()
	} else {
		panic("wrong value of -db flag")
	}
	srv := handler.NewDefaultServer(graphQL.NewExecutableSchema(graphQL.Config{
		Resolvers: &graphQL.Resolver{Db: storage},
	}))
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	})
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	go log.Fatal(http.ListenAndServe(":"+port, nil))
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	err := storage.Stop()
	if err != nil {
		println(err.Error())
	}

}
