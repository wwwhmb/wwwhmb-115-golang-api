package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func NewPost(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Message string
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	session := ctx.Value("Session").(*Session)
	if session == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	newPostSQL := "insert into posts (user_id, message) values ($1, $2)"
	_, err = db.ExecContext(ctx, newPostSQL, session.UserID, args.Message)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
