package pg

import (
	"context"
	"fmt"
)

func Migrate(ctx context.Context, c Conn) error {
	return nil
}

var migrations = []migration{
	{schemaNotExists("1.2022.05.12.1"), createSchema},
	{schemaVersionLessThan("1.2022.05.12.1"), nil},
}

func createSchema(ctx context.Context, c Conn) error {
	query := `CREATE SCHEMA incremental`
	if _, err := c.ExecContext(ctx, query); err != nil {
		return err
	}
	return nil
}

type filter func(context.Context, Conn) (bool, error)
type action func(context.Context, Conn) error

type migration struct {
	filter
	action
}

func schemaNotExists(schema string) filter {
	return func(ctx context.Context, c Conn) (bool, error) {
		query := `SELECT 1 FROM information_schema.schemata WHERE schema_name = $1`
		res, err := c.QueryContext(ctx, query, schema)
		if err != nil {
			return false, err
		}
		defer res.Close()
		return !res.Next(), nil
	}
}

func schemaVersionLessThan(schemaVersion string) filter {
	return func(ctx context.Context, c Conn) (bool, error) {
		currentSchemaVersion, err := readSchemaVersion(ctx, c)
		if err != nil {
			return false, err
		}
		return schemaVersion < currentSchemaVersion, nil
	}
}

func readSchemaVersion(ctx context.Context, c Conn) (string, error) {
	query := `select version from incremental.schema_version`
	res, err := c.QueryContext(ctx, query)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Close() }()
	var sv schemaVersion
	if !res.Next() {
		return "", fmt.Errorf("no schema version row present in db")
	}
	if err := sv.Populate(res); err != nil {
		return "", err
	}
	return sv.Version, nil
}
