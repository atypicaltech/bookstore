package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	vault "github.com/hashicorp/vault/api"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

const (
	PORT             = "PORT"
	VAULT_ADDR       = "VAULT_ADDR"
	VAULT_TOKEN_FILE = "VAULT_TOKEN_FILE"

	DB_HOST = "DB_HOST"
	DB_PORT = "DB_PORT"
	DB_NAME = "DB_NAME"
	DB_USER = "DB_USER"
	DB_PASS = "DB_PASS"
)

var (
	conf *viper.Viper
)

func init() {
	conf = viper.New()
	conf.AutomaticEnv()

	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	vaultTokenFile := conf.GetString(VAULT_TOKEN_FILE)
	if _, err := os.Stat(vaultTokenFile); err == nil {
		tokenBytes, err := os.ReadFile(vaultTokenFile)
		if err != nil {
			log.Fatalf("problem reading vault token: %v", err)
		}
		client.SetToken(string(tokenBytes))
	}

	secret, err := client.KVv2("internal").Get(context.Background(), "bookstore/env")
	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}

	err = conf.MergeConfigMap(secret.Data)
	if err != nil {
		log.Fatalf("unable to merge secret: %v", err)
	}
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
