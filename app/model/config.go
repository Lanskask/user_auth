package model

type DBConfig struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	sslmode  string
}

func NewDefaultDBConfig() DBConfig {
	return DBConfig{
		host:     "localhost",
		port:     "5432",
		user:     "postgres",
		password: "postgres",
		dbname:   "postgres",
		sslmode:  "disable",
	}
}
