package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type HTTPResponse struct {
	Success bool
	Code    int
	Message string
}

func InitHandlers(config *Config, db *Database) *httprouter.Router {
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
			Debug("POST request received", "table", table.Name)
			handlePost(w, r, table)
		})

		Routes = append(Routes, Route{"PUT", "/" + table.Name})
		router.PUT("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			Debug("POST request received", "table", table.Name)
			handlePost(w, r, table)
		})
	}

	Debug("Handlers initialized")

	return router
}

func handlePost(w http.ResponseWriter, r *http.Request, table Table) {
	contentType := r.Header.Get("Content-Type")
	Debug("Content-Type is " + contentType)
	w.Header().Set("Content-Type", "application/json")

	switch strings.ToLower(contentType) {
	case "application/json":
		var jsonData interface{}
		err := json.NewDecoder(r.Body).Decode(&jsonData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch data := jsonData.(type) {
		case map[string]interface{}:
			Debug("JSON object received")
			// handle JSON object
			response, newEntity := handleJsonObject(data, &table)
			if !response.Success {
				http.Error(w, response.Message, response.Code)
				return
			}
			err = json.NewEncoder(w).Encode(newEntity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		case []interface{}:
			// handle JSON array
			// you can iterate over the array with a for loop
		default:
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		_, _ = w.Write([]byte("POST request received"))
	default:
		http.Error(w, "Invalid content type", http.StatusBadRequest)
	}
}

func handleJsonObject(data Entity, table *Table) (HTTPResponse, *Entity) {
	entity := Entity{}

	// iterate over the JSON object and validate fields
	for key, value := range data {
		Debug("Validating field", "field", key, "value", value)
		if field, ok := table.Definition[key]; ok {
			validation := ValidateField(field, value, key, table)
			Debug("Validation result", "valid", validation.Valid, "errors", validation.Errors)
			if validation.Valid {
				entity[key] = value
				if field.Type == "id" && field.Subtype != "uuid" {
					if uint(value.(float64)) > table.LastAutoID {
						table.LastAutoID = uint(value.(float64)) + 1
					}
				}
			} else {
				return HTTPResponse{false, http.StatusBadRequest, validation.Errors[0]}, nil
			}
		} else {
			return HTTPResponse{false, http.StatusUnprocessableEntity, "Unknown field: " + key}, nil
		}
	}

	// check if all required fields are present and generate missing optional fields
	for key, field := range table.Definition {
		if _, ok := entity[key]; !ok {
			if field.Required {
				return HTTPResponse{false, http.StatusBadRequest, "Missing required field: " + key}, nil
			}

			entity[key], table = GenerateEntityField(*field, table)
		}
	}

	err := AppendTable(table, entity)

	if err != nil {
		return HTTPResponse{false, http.StatusInternalServerError, "Error creating entity"}, nil
	}

	return HTTPResponse{true, http.StatusOK, "Entity created!"}, &entity
}
