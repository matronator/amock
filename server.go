package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func InitHandlers(config *Config, db Database) *httprouter.Router {
	fmt.Println("Initializing handlers...")
	router := httprouter.New()

	for _, table := range db.Tables {
		router.GET("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			content, err := GetTable(&table)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			_, _ = w.Write(content)
		})

		router.POST("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			// var body []byte
		})
	}

	return router
}
