package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jwalton/gchalk"
)

type Database struct {
	Tables map[string]Table
}

type Table struct {
	Name           string
	File           string
	DefinitionFile string
	Definition     map[string]*Field
	SchemaFile     string
	LastAutoID     uint
}

type Entity map[string]any

type EntityCollection []Entity

type EntityJSON map[string]string

type Filter struct {
	Field    string
	Operator string
	Value    any
	Apply    func(EntityCollection) EntityCollection
}

type Sort struct {
	Field string
	Order string
}

type PaginatedItems struct {
	First int
	Last  int
	Prev  int
	Next  int
	Pages int
	Count int
	Items EntityCollection
}

type EntityIds map[string]uint

func GenerateEntity(entity EntityJSON, table *Table) (Entity, *Table) {
	fields := make(Entity, len(entity))

	for key, value := range entity {
		options := FieldOptions{false, false, false}
		fieldName := key

		if strings.HasSuffix(key, "!") {
			options.Required = true
			fieldName = strings.TrimSuffix(key, "!")
		} else if strings.HasSuffix(key, "?") {
			options.Nullable = true
			fieldName = strings.TrimSuffix(key, "?")
		} else if strings.HasSuffix(key, "[]") {
			options.Children = true
			fieldName = strings.TrimSuffix(key, "[]")
		}
		fields[fieldName], table = GenerateField(fieldName, value, table, options)
	}

	return fields, table
}

func HydrateDatabase(db *Database) *Database {
	now := time.Now()
	Debug("Building database...")

	var entityJSON EntityJSON

	for key, table := range db.Tables {
		updated := CreateTable(&table, entityJSON)
		db.Tables[key] = *updated
	}

	elapsed := time.Since(now).String()
	Debug("Database is ready! " + gchalk.Italic("("+elapsed+")"))

	return db
}

func CreateTable(table *Table, entityJSON EntityJSON) *Table {
	filename := table.Name + ".amock.json"
	dir := path.Join(DataDir, filename)
	schemaDir := path.Join(SchemaDir, table.Name+".amock.schema.json")

	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		if _, err = os.Stat(schemaDir); !errors.Is(err, os.ErrNotExist) {
			Debug("Table "+gchalk.Bold(table.Name)+" found at "+gchalk.Italic(dir)+" - skipping...", "table", table.Name, "file", dir, "schema", schemaDir)
			table.File = dir
			table.SchemaFile = schemaDir

			var schema []byte
			schema, err = os.ReadFile(table.SchemaFile)
			if err != nil {
				log.Fatal(err)
			}

			err = json.Unmarshal(schema, &table.Definition)
			if err != nil {
				log.Fatal(err)
			}

			return table
		}
	}

	raw, err := os.ReadFile(table.DefinitionFile)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(raw, &entityJSON)
	if err != nil {
		log.Fatal(err)
	}

	entities := make([]Entity, config.InitCount)

	for i := 0; i < config.InitCount; i++ {
		entities[i], table = GenerateEntity(entityJSON, table)
	}

	schema, _ := json.Marshal(table.Definition)
	_ = os.WriteFile(schemaDir, schema, os.ModePerm)

	b, _ := json.Marshal(entities)

	_ = os.WriteFile(dir, b, os.ModePerm)

	table.File = dir

	Debug("Table "+gchalk.Bold(table.Name)+" created at "+gchalk.Italic(dir)+" from file "+gchalk.Bold(table.DefinitionFile), "table", table.Name, "file", dir, "schema", table.DefinitionFile)

	return table
}

// func SearchTable(table *Table, filters map[string]any) (EntityCollection, error) {
//
// }

func GetTable(table *Table) ([]byte, error) {
	raw, err := os.ReadFile(table.File)

	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", table.File, err)
	}

	return raw, err
}

func GetEntityById(table *Table, id string) ([]byte, error) {
	collection, err := ReadTable(table)
	if err != nil {
		return nil, err
	}

	var (
		entity *Entity
		found  bool
	)

	newId, err := strconv.ParseFloat(id, 64)
	if err != nil {
		entity, found = FindBy[string](&collection, "id", id)
	} else {
		entity, found = FindBy[float64](&collection, "id", newId)
	}

	if !found {
		return nil, errors.New("entity not found, id: " + id)
	}

	b, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("could not marshal entity: %w", err)
	}

	return b, err
}

func FindBy[T comparable](collection *EntityCollection, key string, search T) (*Entity, bool) {
	for _, entity := range *collection {
		Debug("Comparing", "key", key, "search", search, "entity", entity[key])
		if entity[key] == search {
			Debug("Found entity", "entity", entity)
			return &entity, true
		}
	}

	return nil, false
}

func ReadTable(table *Table) (EntityCollection, error) {
	raw, err := GetTable(table)

	if err != nil {
		return nil, err
	}

	var collection EntityCollection

	err = json.Unmarshal(raw, &collection)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal file %s: %w", table.File, err)
	}

	return collection, err
}

func WriteTable(table *Table, collection EntityCollection) error {
	b, err := json.Marshal(collection)

	if err != nil {
		return fmt.Errorf("could not marshal collection: %w", err)
	}

	err = os.WriteFile(table.File, b, os.ModePerm)

	return err
}

func AppendTable(table *Table, entity *Entity) error {
	collection, err := ReadTable(table)

	if err != nil {
		return err
	}

	collection = append(collection, *entity)

	err = WriteTable(table, collection)

	return err
}

func RemoveById(table *Table, id string) error {
	collection, err := ReadTable(table)

	Debug("Removing entity", "id", id, "table", table.Name)

	if err != nil {
		return err
	}

	for i, entity := range collection {
		if entity["id"] == id {
			collection = append(collection[:i], collection[i+1:]...)
			break
		}
	}

	err = WriteTable(table, collection)

	if err != nil {
		return err
	}

	return nil
}
