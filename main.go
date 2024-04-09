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
	"github.com/jwalton/gchalk"
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
var SchemaDir = path.Join(".amock", "schema")

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

type Config struct {
	Host      string   `yaml:"host" env:"AMOCK_HOST" env-default:"localhost"`
	Port      int      `yaml:"port" env:"AMOCK_PORT" env-default:"8080"`
	Dir       string   `yaml:"dir" env:"AMOCK_DIR"`
	Entities  []string `yaml:"entities" env:"AMOCK_ENTITIES"`
	InitCount int      `yaml:"initCount" env:"AMOCK_INIT_COUNT" env-default:"20"`
}

type Entity map[string]any

type EntityCollection []Entity

type EntityJSON map[string]string

var config *Config

var db Database

func init() {
	InitLogger()
	Debug("Creating database from config...")

	config, _ = ParseConfigFiles(ConfigPaths...)

	if config == nil {
		log.Fatal("No configuration file found")
	}

	buildTablesFromConfig()

	if _, err := os.Stat(DataDir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(DataDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(SchemaDir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(SchemaDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	Debug("Database created")

	db = *HydrateDatabase(&db)
}

func main() {
	StartServer()
}

func StartServer() {
	url := constructUrl()

	fmt.Println(gchalk.Bold("Starting server at " + url))
	fmt.Println("\nAvailable routes:")

	router := InitHandlers(config, &db)
	for _, route := range Routes {
		fmt.Println("  " + gchalk.Bold(RequestMethodColor(route.Method, false)) + "\t" + url + route.Path + "\t" + gchalk.Dim("[entity: "+gchalk.WithItalic().Bold(strings.Split(route.Path, "/")[1])+"]"))
	}
	fmt.Println("")

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

func buildTablesFromConfig() {
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
					Name:           name,
					DefinitionFile: path.Join(config.Dir, filename),
					Definition:     make(map[string]*Field),
					File:           filename,
					LastAutoID:     1,
				}
			}
		}
	}

	if len(config.Entities) > 0 {
		for _, entity := range config.Entities {
			if path.Ext(entity) == ".json" {
				name := entity[:len(entity)-5]
				db.Tables[name] = Table{
					Name:           name,
					DefinitionFile: entity,
					Definition:     make(map[string]*Field),
					File:           path.Base(entity),
					LastAutoID:     1,
				}
			}
		}
	}
}

func constructUrl() string {
	var url string
	if strings.Contains(config.Host, "http://") || strings.Contains(config.Host, "https://") {
		url = config.Host + ":" + strconv.Itoa(config.Port)
	} else {
		url = "http://" + config.Host + ":" + strconv.Itoa(config.Port)
	}
	return url
}
