package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"bookstore.atypicaltech.dev/models"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// This time make models.BookModel the dependency in Env.
type Env struct {
	books interface {
		All() ([]models.Book, error)
	}
}

const (
	PORT    = "PORT"
	DB_HOST = "DB_HOST"
	DB_PORT = "DB_PORT"
	DB_NAME = "DB_NAME"
	DB_USER = "DB_USER"
	DB_PASS = "DB_PASS"
)

var conf *viper.Viper

func init() {
	conf = viper.New()
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
		books: models.BookModel{DB: db},
	}

	http.HandleFunc("/books", env.booksIndex)
	http.ListenAndServe(fmt.Sprintf(":%s", conf.Get(PORT)), nil)
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
