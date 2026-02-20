package store

import (
	"database/sql"
	"io/fs"
	"sort"

	"github.com/aellingwood/cielo/migrations"
)

func RunMigrations(db *sql.DB) error {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		data, err := fs.ReadFile(migrations.FS, f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(data)); err != nil {
			return err
		}
	}
	return nil
}
