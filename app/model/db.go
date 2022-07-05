package model

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

type IDB interface {
	CreateUser(user *User) (string, error)
	GetUser(id string) (User, error)
	DeleteUser(id string) error
	CheckEmailExistence(email string, user *User) (bool, error)
	TruncateTable() error
	Close() error
}

var currentDB IDB // for singleton

func NewDB(conf DBConfig) (IDB, error) {
	if currentDB != nil {
		return currentDB, nil
	}

	dsn := DNSByConfig(conf)
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return &DB{}, fmt.Errorf("error opening a DB: %s", err)
	}

	if err := db.Ping(); err != nil {
		return &DB{}, fmt.Errorf("cann't ping a DB: %s", err)
	}

	currentDB = &DB{db: db}
	return currentDB, nil
}

func (db *DB) CreateUser(user *User) (string, error) {
	var id uint64
	err := db.db.QueryRow(
		"INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id", user.Name, user.Email, user.Password,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("error executing a query: %s", err)
	}

	return strconv.Itoa(int(id)), nil
}

func (db *DB) GetUser(id string) (User, error) {
	rows, err := db.db.Query("select * from users where id=$1", id)
	if err != nil {
		return User{}, fmt.Errorf("error queriing a user from db: %s", err)
	}

	var res User
	for rows.Next() {
		if err := rows.Scan(&res.ID, &res.Name, &res.Email, &res.Password); err != nil {
			return User{}, fmt.Errorf("error scanning wors to get a user: %s", err)
		}
	}

	return res, nil
}

func (db *DB) DeleteUser(id string) error {
	_, err := db.db.Exec("delete from users where id = $1", id)
	if err != nil {
		return fmt.Errorf("err executing a delete: %s", err)
	}
	return nil
}

func (db *DB) CheckEmailExistence(email string, user *User) (bool, error) {
	rows, err := db.db.Query("select id, name, email, password from users where email = $1", email)
	if err != nil {
		return false, fmt.Errorf("err selecting by email: %s", err)
	}
	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
			return false, fmt.Errorf("error scanning rows: %s", err)
		}
	}
	return true, nil
}

func (db *DB) TruncateTable() error {
	_, err := db.db.Exec("TRUNCATE users")
	if err != nil {
		return fmt.Errorf("error truncating users table")
	}
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func DNSByConfig(conf DBConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		conf.host,
		conf.port,
		conf.user,
		conf.password,
		conf.dbname,
		conf.sslmode,
	)
}
