package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type App struct {
	DB    *pgx.Conn
	Redis *redis.Client
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	POSTGRES_CONN := os.Getenv("POSTGRES_CONN")
	REDIS_CONN := os.Getenv("REDIS_CONN")

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, POSTGRES_CONN)
	if err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}

	fmt.Println("Connecting to DB:", POSTGRES_CONN)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     REDIS_CONN,
		Password: "",
		DB:       0,
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Cannot connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis")

	defer redisClient.Close()

	defer conn.Close(ctx)

	app := &App{
		DB: conn,
		Redis: redisClient,
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	r.Post("/register", app.registerUser);

	r.Post("/login", app.login);

	fmt.Println("connecting to localhost:8080")

	http.ListenAndServe(":8080", r)

}

func( a * App) registerUser(w http.ResponseWriter, r *http.Request){
   // validate the request

   
}

func( a * App) login(w http.ResponseWriter, r *http.Request){

}
