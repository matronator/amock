package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func InitHandlers(config *Config, db Database) *httprouter.Router {
	Debug("Initializing handlers...")
	router := httprouter.New()

	for _, table := range db.Tables {
		Routes = append(Routes, Route{"GET", "/" + table.Name})
		router.GET("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			content, err := GetTable(&table)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			_, _ = w.Write(content)
		})

		Routes = append(Routes, Route{"POST", "/" + table.Name})
		router.POST("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			// var body []byte
		})
	}

	Debug("Handlers initialized")

	return router
}
