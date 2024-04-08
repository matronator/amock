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

func GetTable(table *Table) ([]byte, error) {
	raw, err := os.ReadFile(table.File)

	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", table.File, err)
	}

	return raw, err
}

func GetEntity(table *Table, id string) ([]byte, error) {
	collection, err := ReadTable(table)

	if err != nil {
		return nil, err
	}

	for _, entity := range collection {
		newId, _ := strconv.ParseFloat(id, 64)
		if entity["id"] == newId {
			b, err2 := json.Marshal(entity)
			if err2 != nil {
				return nil, fmt.Errorf("could not marshal entity: %w", err2)
			}

			return b, err
		}
	}

	return nil, errors.New("entity not found")

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

func AppendTable(table *Table, entity Entity) error {
	collection, err := ReadTable(table)

	if err != nil {
		return err
	}

	collection = append(collection, entity)

	err = WriteTable(table, collection)

	return err
}
