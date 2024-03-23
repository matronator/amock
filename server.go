package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func InitHandlers(config *Config, db Database) {
	router := httprouter.New()

	for _, table := range db.Tables {
		router.GET("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

			_, _ = w.Write([]byte("Hello, " + ps.ByName("name") + "!"))
		})
	}
}
