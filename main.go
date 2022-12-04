package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

const (
	PORT = "PORT"

	VAULT_ADDR          = "VAULT_ADDR"
	VAULT_ROLE          = "VAULT_ROLE"
	VAULT_KV_MOUNT      = "VAULT_KV_MOUNT"
	VAULT_BOOKSTORE_ENV = "VAULT_BOOKSTORE_ENV"

	KUBE_SVC_ACCT_TOKEN = "KUBE_SVC_ACCT_TOKEN"

	DB_HOST = "DB_HOST"
	DB_PORT = "DB_PORT"
	DB_NAME = "DB_NAME"
	DB_USER = "DB_USER"
	DB_PASS = "DB_PASS"
	DB_SSL  = "DB_SSL"
)

var (
	conf *viper.Viper
)

func init() {
	conf = viper.New()
	conf.AutomaticEnv()

	kvMount := conf.GetString(VAULT_KV_MOUNT)
	bookstoreEnv := conf.GetString(VAULT_BOOKSTORE_ENV)

	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	err = loginVaultKubernetes(client)
	if err != nil {
		log.Println("vault login failed: %w", err)
	}

	secret, err := client.KVv2(kvMount).Get(context.Background(), bookstoreEnv)
	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}

	err = conf.MergeConfigMap(secret.Data)
	if err != nil {
		log.Fatalf("unable to merge secret: %v", err)
	}
}

func main() {
	port := conf.GetString(PORT)

	dbUser := conf.GetString(DB_USER)
	dbPass := conf.GetString(DB_PASS)
	dbHost := conf.GetString(DB_HOST)
	dbPort := conf.GetString(DB_PORT)
	dbName := conf.GetString(DB_NAME)
	dbSSL := conf.GetString(DB_SSL)

	dataSourceName := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPass, dbHost, dbPort, dbName, dbSSL)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	env := &Env{
		books: BookModel{DB: db},
	}

	http.HandleFunc("/books", env.booksIndex)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
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

func loginVaultKubernetes(client *vault.Client) error {
	vaultRole := conf.GetString(VAULT_ROLE)
	kubeToken := conf.GetString(KUBE_SVC_ACCT_TOKEN)

	k8sAuth, err := auth.NewKubernetesAuth(
		vaultRole,
		auth.WithServiceAccountTokenPath(kubeToken),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), k8sAuth)
	if err != nil {
		return fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
	}
	if authInfo == nil {
		return fmt.Errorf("no auth info was returned after login")
	}

	return nil
}
