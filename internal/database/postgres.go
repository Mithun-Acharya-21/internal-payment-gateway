package database

import (
	"os"
	"path/filepath"
	"strings"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgres(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	return db, nil
}

func Migrate(db *sqlx.DB) error {
	path := filepath.Join("migrations", "001_init_schema.sql")
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(sqlBytes), "\n")
	withoutComments := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		withoutComments = append(withoutComments, line)
	}
	queries := strings.Split(strings.Join(withoutComments, "\n"), ";")
	for _, q := range queries {
		query := strings.TrimSpace(q)
		if query == "" || strings.HasPrefix(query, "--") {
			continue
		}
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

