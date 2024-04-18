package redshift

import (
	"fmt"

	"github.com/artie-labs/transfer/lib/config/constants"
	"github.com/artie-labs/transfer/lib/sql"
)

type TableIdentifier struct {
	schema string
	table  string
}

func NewTableIdentifier(schema, table string) TableIdentifier {
	return TableIdentifier{schema: schema, table: table}
}

func (ti TableIdentifier) Schema() string {
	return ti.schema
}

func (ti TableIdentifier) Table() string {
	return ti.table
}

func (ti TableIdentifier) FullyQualifiedName(escape, uppercaseEscNames bool) string {
	// Redshift is Postgres compatible, so when establishing a connection, we'll specify a database.
	// Thus, we only need to specify schema and table name here.
	return fmt.Sprintf(
		"%s.%s",
		ti.schema,
		sql.EscapeName(ti.table, uppercaseEscNames, &sql.NameArgs{Escape: escape, DestKind: constants.Redshift}),
	)
}