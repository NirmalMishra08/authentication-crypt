package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"main.go/db"
)

type App struct {
	DB      *pgx.Conn
	Redis   *redis.Client
	Queries *db.Queries
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	POSTGRES_CONN := os.Getenv("POSTGRES_CONN")

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, POSTGRES_CONN)
	if err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}

	fmt.Println("Connecting to DB:", POSTGRES_CONN)

	opt, err := redis.ParseURL(os.Getenv("REDIS_CONN"))
	if err != nil {
		log.Fatalf("Redis parse error: %v", err)
	}

	redisClient := redis.NewClient(opt)

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Cannot connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis")

	defer redisClient.Close()

	defer conn.Close(ctx)

	queries := db.New(conn)

	app := &App{
		DB:      conn,
		Redis:   redisClient,
		Queries: queries,
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	r.Post("/register", app.registerUser)

	r.Post("/login", app.login)

	fmt.Println("connecting to localhost:8080")

	http.ListenAndServe(":8080", r)

}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) registerUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// validate the request
	var data registerRequest

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "not able to generate password", 400)
		return
	}

	err = a.Queries.InsertUser(ctx, db.InsertUserParams{
		Username: data.Username,
		Password: pgtype.Text{String: string(hash), Valid: true},
	})

	fmt.Println(data.Username, data.Password)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.WriteHeader(201)
	w.Write([]byte("created new user"))

}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// validate the request
	var data registerRequest

	ip := GetIP(r)

	ipAttempts, _ := a.IncrementIPAttempts(ip)

	if ipAttempts > 20 {
		fmt.Errorf("too many requests")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	user, err := a.Queries.GetUserByUserName(ctx, data.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Prevent user enumeration
			time.Sleep(500 * time.Millisecond)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// real DB error
		http.Error(w, "database error", 500)
		return
	}

	locked, err := a.IsUserLocked(strconv.Itoa(int(user.ID)))
	if locked {
		fmt.Errorf("account locked, try later")
		return
	}

	userIdStr := strconv.Itoa(int(user.ID))

	err = bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(data.Password))
	if err != nil {
		attempts, _ := a.IncrementAttempts(userIdStr)
		// progressive delay
		delay := time.Duration(attempts*500) * time.Millisecond
		if delay > 5*time.Second {
			delay = 5 * time.Second
		}
		time.Sleep(delay)

		// lock if too many attempts
		if attempts >= 5 {
			a.LockUser(userIdStr)
		}

		fmt.Errorf("invalid email or password")
		return

	}

	w.Write([]byte("Login successful"))

}

func (a *App) IsUserLocked(userID string) (bool, error) {
	ctx := context.Background()
	key := "login:lock:user:" + userID

	val, err := a.Redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return val == 1, nil
}

func (a *App) IncrementAttempts(userID string) (int64, error) {
	ctx := context.Background()
	key := "login:attempts:user:" + userID

	// increment
	count, err := a.Redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		a.Redis.Expire(ctx, key, 15*time.Minute)
	}

	return count, nil
}

func (a *App) LockUser(userID string) error {
	ctx := context.Background()
	key := "login:lock:user:" + userID

	return a.Redis.Set(ctx, key, "1", 15*time.Minute).Err()
}

func (a *App) ResetAttempts(userID string) {
	ctx := context.Background()
	a.Redis.Del(ctx, "login:attempts:user:"+userID)
	a.Redis.Del(ctx, "login:lock:user:"+userID)
}

func (a *App) IncrementIPAttempts(ip string) (int64, error) {
	context := context.Background()
	key := "login:attempts:ip:" + ip

	count, err := a.Redis.Incr(context, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		a.Redis.Expire(context, key, 15*time.Minute)
	}

	return count, nil
}

func GetIP(r *http.Request) string {
	ip := r.RemoteAddr
	return ip
}
