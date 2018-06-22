// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.

package webserver

import (
	"errors"
	"html/template"
	"net/http"
	"time"
)

type WebHandle func(w http.ResponseWriter, r *http.Request)

type webserver struct {
	url2handle map[string]WebHandle
}

func (this *webserver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error = nil
	path := r.URL.String()
	switch path {
	case "/":
		t := template.Must(template.ParseFiles("../webserver/root.html"))
		t.Execute(w, time.Now().Format("2006-01-02 15:04:05"))

	default:
		err = errors.New("unrecognized transaction type")
	}

	if err != nil {
		http.Error(w, "error 500: "+err.Error(), http.StatusInternalServerError)
	}
}

func StartWebServer() {
	http.ListenAndServe(":8080", &webserver{})
}
