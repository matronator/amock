package main

import (
	"encoding/json"
	"errors"
	"flag"
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
var TablesDir = path.Join(".amock", "tables")

type Config struct {
	Host      string   `yaml:"host" env:"AMOCK_HOST" env-default:"localhost"`
	Port      int      `yaml:"port" env:"AMOCK_PORT" env-default:"8080"`
	Dir       string   `yaml:"dir" env:"AMOCK_DIR"`
	Entities  []string `yaml:"entities" env:"AMOCK_ENTITIES"`
	InitCount int      `yaml:"initCount" env:"AMOCK_INIT_COUNT" env-default:"20"`
}

var config *Config

var db Database

func init() {
	InitLogger()
	Debug("Creating database from config...")

	config, _ = parseConfigFiles(ConfigPaths...)

	if config == nil {
		log.Fatal("No configuration file found")
	}

	getHostFromArgs()

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

func getHostFromArgs() {
	host := flag.Arg(0)
	if host != "" {
		var noPrefix string
		var prefix string
		if strings.Contains(host, "http://") {
			noPrefix = strings.TrimPrefix(host, "http://")
			prefix = "http://"
		} else if strings.Contains(host, "https://") {
			noPrefix = strings.TrimPrefix(host, "https://")
			prefix = "https://"
		} else {
			noPrefix = host
		}
		if strings.Contains(noPrefix, ":") {
			parts := strings.Split(noPrefix, ":")
			config.Host = prefix + parts[0]

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Fatal(err)
			}
			config.Port = port
		} else {
			config.Host = host
		}
	}
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

func parseConfigFiles(files ...string) (*Config, error) {
	var cfg Config
	fileRead := false

	for i := 0; i < len(files); i++ {
		if _, err := os.Stat(files[i]); errors.Is(err, os.ErrNotExist) {
			continue
		}

		err := cleanenv.ReadConfig(files[i], &cfg)
		if err == nil {
			fileRead = true
		}
	}

	if !fileRead {
		err := cleanenv.ReadEnv(&cfg)
		if err != nil {
			log.Printf("Error reading configuration from file or environment: %v\n", err)
			return nil, err
		}
	}

	return &cfg, nil
}

func buildTablesFromConfig() {
	if _, err := os.Stat(TablesDir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(TablesDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	db.Tables = make(map[string]Table)

	if config.Dir != "" {
		dir, err := os.ReadDir(config.Dir)

		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range dir {
			filename := entry.Name()
			table, name := getOrCreateTable(filename, path.Join(config.Dir, filename))
			db.Tables[name] = *table
		}
	}

	if len(config.Entities) > 0 {
		for _, entity := range config.Entities {
			table, name := getOrCreateTable(entity, entity)
			db.Tables[name] = *table
		}
	}
}

func getOrCreateTable(filename string, definitionFile string) (*Table, string) {
	createNew := false
	tempTable := Table{}
	var name string

	if path.Ext(filename) == ".json" {
		tableFilePath := path.Join(TablesDir, filename+".table")
		name = strings.ToLower(filename[:len(filename)-5])

		if _, err := os.Stat(tableFilePath); errors.Is(err, os.ErrNotExist) {
			createNew = true
		} else {
			tableFile, err := os.ReadFile(tableFilePath)
			if err != nil {
				createNew = true
			}

			Debug("Table "+gchalk.Bold(name)+" found at "+gchalk.Italic(tableFilePath)+" - skipping...", "table", name, "file", tableFilePath)

			err = json.Unmarshal(tableFile, &tempTable)
			if err != nil {
				createNew = true
			}
		}
	}

	if createNew {
		tempTable = createNewTable(name, filename, definitionFile)
	}

	return &tempTable, name
}

func createNewTable(name string, filename string, definitionFile string) Table {
	return Table{
		Name:           name,
		DefinitionFile: definitionFile,
		Definition:     make(map[string]*Field),
		File:           filename,
		LastAutoID:     1,
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
