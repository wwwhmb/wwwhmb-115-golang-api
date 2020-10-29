package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Username string
		Password string
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	registerSQL := `
		insert into users (username, password)
			values ($1, crypt($2, gen_salt('bf')))`
	_, err = db.ExecContext(ctx, registerSQL, args.Username, args.Password)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Username string
		Password string
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	loginSQL := `
		insert into sessions (user_id)
			select user_id
			from users
			where username = $1
			and password = crypt($2, password)
		returning session_id`
	row := db.QueryRowContext(ctx, loginSQL, args.Username, args.Password)
	var sessionID string
	err = row.Scan(&sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sessionID))
}

type Post struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func Posts(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	username := params.Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	postsSQL := `
		select u.username, p.message, p.created_at
		from users u
		inner join posts p
		  on u.user_id = p.user_id
		where u.username = $1
		order by created_at desc`
	rows, err := db.QueryContext(ctx, postsSQL, username)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	posts := []Post{}
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Username, &post.Message, &post.CreatedAt)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	postsJSON, err := json.Marshal(posts)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(postsJSON)
}
