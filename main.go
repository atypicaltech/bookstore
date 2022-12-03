package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

const (
	SECRETS_FILE = "SECRETS_FILE"
	PORT         = "PORT"
	DB_HOST      = "DB_HOST"
	DB_PORT      = "DB_PORT"
	DB_NAME      = "DB_NAME"
	DB_USER      = "DB_USER"
	DB_PASS      = "DB_PASS"
)

var (
	conf       *viper.Viper
	configPath string
	configName string
)

func init() {
	conf = viper.New()

	if usingVaultSecrets() {
		conf.AddConfigPath(configPath)
		conf.SetConfigName(configName)
		err := conf.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	conf.AutomaticEnv()
}

func main() {
	dataSourceName := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		conf.Get(DB_USER),
		conf.Get(DB_PASS),
		conf.Get(DB_HOST),
		conf.Get(DB_NAME),
	)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	env := &Env{
		books: BookModel{DB: db},
	}

	http.HandleFunc("/books", env.booksIndex)
	http.ListenAndServe(fmt.Sprintf(":%s", conf.Get(PORT)), nil)
}

type Env struct {
	books interface {
		All() ([]Book, error)
	}
}

func (env *Env) booksIndex(w http.ResponseWriter, r *http.Request) {
	// Execute the SQL query by calling the All() method.
	bks, err := env.books.All()
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, bk := range bks {
		fmt.Fprintf(w, "%s, %s, %s, $%.2f\n", bk.Isbn, bk.Title, bk.Author, bk.Price)
	}
}

type Book struct {
	Isbn   string
	Title  string
	Author string
	Price  float32
}

// Create a custom BookModel type which wraps the sql.DB connection pool.
type BookModel struct {
	DB *sql.DB
}

// Use a method on the custom BookModel type to run the SQL query.
func (m BookModel) All() ([]Book, error) {
	rows, err := m.DB.Query("SELECT * FROM books")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bks []Book

	for rows.Next() {
		var bk Book

		err := rows.Scan(&bk.Isbn, &bk.Title, &bk.Author, &bk.Price)
		if err != nil {
			return nil, err
		}

		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bks, nil
}

func usingVaultSecrets() bool {
	fullSecretsPath := os.Getenv(SECRETS_FILE)
	if fullSecretsPath == "" {
		return false
	}

	secrets, err := filepath.Abs(fullSecretsPath)
	if err != nil {
		return false
	}

	configPath = filepath.Dir(secrets)
	fileName := filepath.Base(secrets)
	configName = strings.TrimSuffix(fileName, filepath.Ext(configName))

	return strings.Contains(configPath, "/vault/secrets")
}
