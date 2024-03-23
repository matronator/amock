package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/timshannon/bolthold"
)

func GenerateEntity(entity EntityJSON, table *Table) (Entity, *Table) {
	fields := make(Entity, len(entity))

	for key, value := range entity {
		fields[key], table = GenerateField(value, table)
	}

	return fields, table
}

func HydrateDatabase(db Database) Database {
	var entityJSON EntityJSON

	store, err := bolthold.Open("amock.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	for key, table := range db.Tables {
		var updated Table
		updated, store = CreateTable(&table, entityJSON, store)
		db.Tables[key] = updated
	}

	return db
}

func CreateTable(table *Table, entityJSON EntityJSON, store *bolthold.Store) (Table, *bolthold.Store) {
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

	filename := table.Name + ".amock.json"

	_ = os.WriteFile(filename, b, os.ModePerm)

	table.File = filename

	return *table, store
}
