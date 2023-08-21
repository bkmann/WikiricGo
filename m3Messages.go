package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"net/http"
)

type ChatMessage struct {
	ChatGroupUUID string `json:"parent"`
	Text          string `json:"msg"`
	Username      string `json:"usr"`
	TimeCreated   string `json:"ts"`
}

func OpenChatMessageDatabase() *GoDB {
	db := OpenDB(
		"chatMessages", []string{
			"chatID",
		},
	)
	return db
}

func (db *GoDB) ProtectedChatMessagesEndpoints(r chi.Router, tokenAuth *jwtauth.JWTAuth, chatServer *ChatServer) {
	r.Route(
		"/msg/private", func(r chi.Router) {
			r.Post("/create", db.handleChatMessageCreate(chatServer))
		},
	)
}

func (db *GoDB) handleChatMessageCreate(chatServer *ChatServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*User)
		if user == nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// Retrieve POST payload
		request := &ChatMessage{}
		if err := render.Bind(r, request); err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// TODO: Check rights
		// Initialize message
		request.TimeCreated = TimeNowIsoString()
		request.Username = user.Username
		// Distribute message
		go chatServer.DistributeChatMessageJSON(request)
		// Store message
		jsonMessage, err := json.Marshal(request)
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		uUID, err := db.Insert(
			jsonMessage, map[string]string{
				"chatID": request.ChatGroupUUID,
			},
		)
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// Respond to client
		_, _ = fmt.Fprintln(w, uUID)
	}
}

func (a *ChatMessage) Bind(_ *http.Request) error {
	if a.Text == "" {
		return errors.New("missing text")
	}
	if a.ChatGroupUUID == "" {
		return errors.New("missing parent")
	}
	return nil
}
