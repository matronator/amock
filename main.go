package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

var ConfigPaths = []string{
	".amock.json",
	".amockrc",
	".amock.json.json",
	".amock.json.yml",
	".amock.json.yaml",
	".amock.json.toml",
	"amock.config",
	"amock.json",
	"amock.yml",
	"amock.yaml",
	"amock.toml",
}

var DataDir = path.Join(".amock", "data")

type Database struct {
	Tables map[string]Table
}

type Table struct {
	Name       string
	File       string
	Definition string
	LastAutoID uint
}

type Config struct {
	Host      string   `yaml:"host" env:"HOST" env-default:"localhost"`
	Port      int      `yaml:"port" env:"PORT" env-default:"8080"`
	Dir       string   `yaml:"dir" env:"DIR"`
	Entities  []string `yaml:"entities" env:"ENTITIES"`
	InitCount int      `yaml:"init_count" env:"INIT_COUNT" env-default:"20"`
}

type Entity map[string]any

type EntityCollection []Entity

type EntityJSON map[string]string

var config *Config

var db Database

func Init() {
	fmt.Println("Creating database from config...")

	config, _ = ParseConfigFiles(ConfigPaths...)

	if config == nil {
		log.Fatal("No configuration file found")
	}

	db.Tables = make(map[string]Table)

	if config.Dir != "" {
		dir, err := os.ReadDir(config.Dir)

		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range dir {
			filename := entry.Name()

			if path.Ext(filename) == ".json" {
				name := strings.ToLower(filename[:len(filename)-5])
				db.Tables[name] = Table{
					Name:       name,
					Definition: path.Join(config.Dir, filename),
					File:       filename,
					LastAutoID: 1,
				}
			}
		}
	}

	if len(config.Entities) > 0 {
		for _, entity := range config.Entities {
			if path.Ext(entity) == ".json" {
				name := entity[:len(entity)-5]
				db.Tables[name] = Table{
					Name:       name,
					Definition: entity,
					File:       path.Base(entity),
					LastAutoID: 1,
				}
			}
		}
	}

	if _, err := os.Stat(DataDir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(DataDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	db = HydrateDatabase(db)
}

func main() {
	Init()
	router := InitHandlers(config, db)

	var url string
	if strings.Contains(config.Host, "http://") || strings.Contains(config.Host, "https://") {
		url = config.Host + ":" + strconv.Itoa(config.Port)
	} else {
		url = "http://" + config.Host + ":" + strconv.Itoa(config.Port)
	}

	fmt.Println("\nStarting server at " + url)
	log.Fatal(http.ListenAndServe(config.Host+":"+strconv.Itoa(config.Port), LogRequest(router)))
}

func ParseConfigFiles(files ...string) (*Config, error) {
	var cfg Config

	for i := 0; i < len(files); i++ {
		if _, err := os.Stat(files[i]); errors.Is(err, os.ErrNotExist) {
			continue
		}

		err := cleanenv.ReadConfig(files[i], &cfg)
		if err != nil {
			log.Printf("Error reading configuration from file:%v", files[i])
			return nil, err
		}
	}

	return &cfg, nil
}
