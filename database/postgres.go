package database

import (
	"database/sql"
	"fmt"
	"os"
)

type Postgres struct {
	DB *sql.DB
	Host string 
	Port int
	User string 
	Password string 
	DBName string 
}

/* connect to db */
func (p *Postgres) Connect() error {
	connStr := fmt.Sprintf(
    "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
    p.Host, p.Port, p.User, p.Password, p.DBName,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connect to PostgreSQL %v\n", err)
		return err
	}
	p.DB = db
	return nil
}

/* health check */ 
func (p *Postgres) Ping() error {
	if p.DB == nil {
		return fmt.Errorf("no database connection")
	}
	return p.DB.Ping()
}