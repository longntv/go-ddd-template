package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadFixtures inserts the rows of every <table>.yml file in dir into the
// table of the same name. Each file is a YAML list of rows keyed by column.
func loadFixtures(ctx context.Context, db *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return fmt.Errorf("glob fixtures: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		table := strings.TrimSuffix(filepath.Base(file), ".yml")
		if err := loadFixtureFile(ctx, db, table, file); err != nil {
			return fmt.Errorf("load fixture %s: %w", file, err)
		}
	}

	return nil
}

func loadFixtureFile(ctx context.Context, db *sql.DB, table, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var rows []map[string]any
	if err := yaml.Unmarshal(content, &rows); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	for i, row := range rows {
		columns := make([]string, 0, len(row))
		for column := range row {
			columns = append(columns, column)
		}
		sort.Strings(columns)

		placeholders := make([]string, len(columns))
		values := make([]any, len(columns))
		for j, column := range columns {
			placeholders[j] = fmt.Sprintf("$%d", j+1)
			values[j] = row[column]
		}

		stmt := fmt.Sprintf(
			"INSERT INTO %q (%s) VALUES (%s)",
			table,
			`"`+strings.Join(columns, `", "`)+`"`,
			strings.Join(placeholders, ", "),
		)
		if _, err := db.ExecContext(ctx, stmt, values...); err != nil {
			return fmt.Errorf("insert row %d: %w", i, err)
		}
	}

	return nil
}
