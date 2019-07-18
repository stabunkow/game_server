package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"game_server/model"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.FormValue("email")
	if err := ValidateEmail(email); err != nil {
		responseJsonError(w, err, http.StatusUnprocessableEntity)
		return
	}

	password := r.FormValue("password")
	if err := ValidatePassword(password); err != nil {
		responseJsonError(w, err, http.StatusUnprocessableEntity)
		return
	}

	usr := model.FindUserByEmail(email)
	if usr == nil {
		err := errors.New("email not exist")
		responseJsonError(w, err, http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.GetPassword()), []byte(password)); err != nil {
		err := errors.New("password wrong")
		responseJsonError(w, err, http.StatusBadRequest)
		return
	}

	// new sid for user
	sid := UniqueId()
	usr.UpdateSid(sid)

	rsp := make(map[string]string)
	rsp["sid"] = sid

	responseJson(w, rsp)
}

func register(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.FormValue("email")
	if err := ValidateEmail(email); err != nil {
		responseJsonError(w, err, http.StatusUnprocessableEntity)
		return
	}

	password := r.FormValue("password")
	if err := ValidatePassword(password); err != nil {
		responseJsonError(w, err, http.StatusUnprocessableEntity)
		return
	}

	usr := model.FindUserByEmail(email)
	if usr != nil {
		err := errors.New("email existed")
		responseJsonError(w, err, http.StatusUnprocessableEntity)
		return
	}

	model.CreateUser(email, password)
	w.WriteHeader(http.StatusNoContent)
}

func responseJson(w http.ResponseWriter, rst interface{}) {
	w.Header().Set("Content-type", "application/json;	charset=utf-8")
	content, _ := json.Marshal(rst)
	fmt.Fprintln(w, string(content))
}

type ErrorBag struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func responseJsonError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-type", "application/json;	charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	errorBag := &ErrorBag{
		code,
		err.Error(),
	}
	content, _ := json.Marshal(errorBag)

	fmt.Fprintln(w, string(content))
}

func responseJsonInternalError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	err = errors.New("server error")

	responseJsonError(w, err, code)
}
