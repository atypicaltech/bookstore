# bookstore

## Container

### Build image (docker)

```sh
./scripts/build_and_push.sh $tag
```

## Database

### Log in to database

Log in to the database to run the following SQL. It can also be done via Adminer.

```sh
# login as posgres to create database, table and user
# this can be done through adminer as well
PGPASSWORD=$POSTGRES_PASSWORD psql --host 127.0.0.1 -U postgres -d postgres -p 5432
```

### Create database, user, and seed

```sql
create database bookstore;
create user bookstoreuser with encrypted password 'bookstorepassword';
grant all privileges on database bookstore to bookstoreuser;
create table books (
    isbn char(14) NOT NULL,
    title varchar(255) NOT NULL,
    author varchar(255) NOT NULL,
    price decimal(5,2) NOT NULL
);
grant select, insert, update, delete on books to bookstoreuser;

alter table books owner to bookstoreuser;
alter table books add primary key (isbn);

insert into books (isbn, title, author, price) values
('978-1503261969', 'Emma', 'Jayne Austen', 9.44),
('978-1505255607', 'The Time Machine', 'H. G. Wells', 5.99),
('978-1503379640', 'The Prince', 'Niccolò Machiavelli', 6.99);
```


## Variables

| Variable | Description | Required? |
|:---------|:-----------:|:---------:|
| PORT | Port to run server on | yes |
| VAULT_ADDR | Address of Vault server for secrets | yes |
| VAULT_ROLE | Vault role to login with | yes |
| VAULT_KV_MOUNT | Vault KV mount containing secrets | yes |
| VAULT_BOOKSTORE_ENV | Path to bookstore env secret | yes |
| KUBE_SVC_ACCT_TOKEN | Path to kubernetes service account token (used to login to Vault as service account) | yes |
| DB_HOST | Database host | yes |
| DB_PORT | Database port | yes |
| DB_NAME | Database name | yes |
| DB_USER | Database user | yes |
| DB_PASS | Database password | yes |
| DB_SSL  | Database SSL option flag | yes |
