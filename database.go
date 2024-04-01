package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/jwalton/gchalk"
)

func GenerateEntity(entity EntityJSON, table *Table) (Entity, *Table) {
	fields := make(Entity, len(entity))

	for key, value := range entity {
		fields[key], table = GenerateField(value, table)
	}

	return fields, table
}

func HydrateDatabase(db Database) Database {
	now := time.Now()

	var entityJSON EntityJSON

	for key, table := range db.Tables {
		updated := CreateTable(&table, entityJSON)
		db.Tables[key] = updated
	}

	elapsed := time.Since(now).String()
	fmt.Println(gchalk.Italic("Database is ready! " + gchalk.Dim("("+elapsed+")")))

	return db
}

func CreateTable(table *Table, entityJSON EntityJSON) Table {
	filename := table.Name + ".amock.json"
	dir := path.Join(DataDir, filename)

	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		table.File = dir
		fmt.Println(gchalk.Dim("Table", gchalk.Bold(table.Name), "found at", gchalk.Italic(dir), "- skipping..."))

		return *table
	}

	raw, err := os.ReadFile(table.Definition)

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

	b, _ := json.Marshal(entities)

	_ = os.WriteFile(dir, b, os.ModePerm)

	table.File = dir

	fmt.Println(gchalk.Dim("Table", gchalk.Bold(table.Name), "created at", gchalk.Italic(dir), "from file", gchalk.Bold(table.Definition)))

	return *table
}

func GetTable(table *Table) ([]byte, error) {
	raw, err := os.ReadFile(table.File)

	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", table.File, err)
	}

	return raw, err
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
