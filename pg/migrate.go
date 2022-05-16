package pg

import (
	"context"
	"fmt"
)

func Migrate(ctx context.Context, c Conn) error {
	return nil
}

var migrations = []migration{
	{schemaLessThan("1.2022.05.12.1"), nil},
}

type versionFilter func(context.Context, Conn) (bool, error)
type action func(context.Context, Conn) error

type migration struct {
	versionFilter
	action
}

func schemaLessThan(schemaVersion string) versionFilter {
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
