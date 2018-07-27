// Copyright 2018 The eballscan Authors
// This file is part of the eballscan.
//
// The eballscan is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The eballscan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the eballscan. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/ecoball/eballscan/data"
	"github.com/ecoball/eballscan/onlooker"
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
		t := template.Must(template.ParseFiles("./root.html"))
		t.Execute(w, data.PrintBlock())

	default:
		err = errors.New("unrecognized transaction type")
	}

	if err != nil {
		http.Error(w, "error 500: "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	go onlooker.Bystander()
	http.ListenAndServe(":8080", &webserver{})
}
