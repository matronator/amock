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

type Route struct {
	Method string
	Path   string
}

var Routes []Route

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

		Routes = append(Routes, Route{"GET", "/" + table.Name + "/:id"})
		router.GET("/"+table.Name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			content, err := GetEntityById(&table, ps.ByName("id"))

			if err != nil {
				if strings.Contains(err.Error(), "entity not found") {
					http.Error(w, "Entity not found", http.StatusNotFound)
					return
				}

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			_, _ = w.Write(content)
		})

		Routes = append(Routes, Route{"POST", "/" + table.Name})
		router.POST("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			Debug("POST request received", "table", table.Name)
			handlePost(w, r, &table)
		})

		Routes = append(Routes, Route{"PUT", "/" + table.Name})
		router.PUT("/"+table.Name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			Debug("PUT request received", "table", table.Name)
			handlePost(w, r, &table)
		})

		Routes = append(Routes, Route{"DELETE", "/" + table.Name + "/:id"})
		router.DELETE("/"+table.Name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			Debug("DELETE request received", "table", table.Name)

			err := RemoveById(&table, ps.ByName("id"))

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			_, _ = w.Write([]byte(`{"message": "Entity removed"}`))
		})
	}

	Debug("Handlers initialized")

	return router
}

func handlePost(w http.ResponseWriter, r *http.Request, table *Table) {
	contentType := r.Header.Get("Content-Type")
	Debug("Content-Type is " + contentType)

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
			response, newEntity, newTable := handleJsonObject(data, table)
			if !response.Success {
				http.Error(w, response.Message, response.Code)
				return
			}

			err = AppendTable(newTable, newEntity)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = json.NewEncoder(w).Encode(&newEntity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			db.Tables[newTable.Name] = *newTable
		case []map[string]interface{}:
			// handle JSON array
			var collection = EntityCollection{}
			newTable := table
			for _, item := range data {
				var response HTTPResponse
				var newEntity *Entity
				response, newEntity, newTable = handleJsonObject(item, table)
				if !response.Success {
					http.Error(w, response.Message, response.Code)
					return
				}

				collection = append(collection, *newEntity)
			}

			backup, _ := ReadTable(newTable)

			for _, entity := range collection {
				err = AppendTable(newTable, &entity)

				if err != nil {
					_ = WriteTable(newTable, backup)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			err = json.NewEncoder(w).Encode(collection)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			db.Tables[newTable.Name] = *newTable
		default:
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Invalid content type", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
}

func handleJsonObject(data Entity, table *Table) (HTTPResponse, *Entity, *Table) {
	var entity, newTable, response = createEntityFromData(data, table)

	if !response.Success {
		return response, nil, nil
	}

	return HTTPResponse{true, http.StatusCreated, "Entity created!"}, entity, newTable
}

func createEntityFromData(data Entity, table *Table) (*Entity, *Table, HTTPResponse) {
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
				return nil, nil, HTTPResponse{false, http.StatusBadRequest, validation.Errors[0]}
			}
		} else {
			return nil, nil, HTTPResponse{false, http.StatusUnprocessableEntity, "Unknown field: " + key}
		}
	}

	// check if all required fields are present and generate missing optional fields
	for key, field := range table.Definition {
		if _, ok := entity[key]; !ok {
			if field.Required {
				return nil, nil, HTTPResponse{false, http.StatusBadRequest, "Missing required field: " + key}
			}

			entity[key], table = GenerateEntityField(*field, table)
		}
	}
	return &entity, table, HTTPResponse{true, http.StatusCreated, "Entity created!"}
}
