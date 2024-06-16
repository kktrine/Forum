package main

import (
	"context"
	"errors"
	"flag"
	"forum/internal/graphQL"
	"forum/internal/storage"
	"forum/internal/storage/memoryDB"
	"forum/internal/storage/postgres"
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
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	docker := flag.Bool("docker", false, "set this if you run program in docker")
	storageFlag := flag.String("db", "sql", "sql or mem")
	flag.Parse()
	var err error
	if !*docker {
		err = godotenv.Load("./env/.env")
	} else {
		err = godotenv.Load("./env/.env_docker")
	}
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	var storageData storage.Storage
	if *storageFlag == "sql" {
		db := os.Getenv("POSTGRES")
		storageData = postgres.New(db)
	} else if *storageFlag == "mem" {
		storageData = memoryDB.New()
	} else {
		panic("wrong value of -db flag")
	}

	srv := handler.NewDefaultServer(graphQL.NewExecutableSchema(graphQL.Config{
		Resolvers: &graphQL.Resolver{Db: storageData},
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

	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = storageData.Stop()
	if err != nil {
		log.Printf("Error stopping storageData: %v", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")

}
