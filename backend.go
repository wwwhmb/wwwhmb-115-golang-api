package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	fmt.Println("Backend Server")

	connStr := os.Getenv("DB_URI")
	if connStr == "" {
		log.Fatalln("Must set DB_URI connection string")
	}
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()

	// anon routes
	mux.Handle("/register", http.HandlerFunc(Register))
	mux.Handle("/login", http.HandlerFunc(Login))
	mux.Handle("/posts", http.HandlerFunc(Posts))

	// user routes
	mux.Handle("/new_post", AuthUser(http.HandlerFunc(NewPost)))

	logMux := LoggingHandler(mux)

	srv := &http.Server{
		Handler:        logMux,
		Addr:           ":8080",
		WriteTimeout:   120 * time.Second,
		ReadTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}

func LoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Println(r.Method, r.URL.EscapedPath(), time.Since(start))
	})
}

type Session struct {
	UserID int64
}

func AuthUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authToken := strings.TrimPrefix(auth, "Token ")

		sessionSQL := `
			select user_id
			from sessions
			where session_id = $1
			and created_at > now() - interval '12 hours'`
		row := db.QueryRowContext(ctx, sessionSQL, authToken)
		var userID int64
		row.Scan(&userID)
		if userID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, "Session", &Session{userID})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
