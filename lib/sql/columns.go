package sql

import (
	"fmt"
	"strings"

	"github.com/artie-labs/transfer/lib/typing"
	"github.com/artie-labs/transfer/lib/typing/columns"
)

func QuoteColumns(cols []columns.Column, dialect Dialect) []string {
	result := make([]string, len(cols))
	for i, col := range cols {
		result[i] = dialect.QuoteIdentifier(col.Name())
	}
	return result
}

// buildColumnsUpdateFragment will parse the columns and then returns a list of strings like: cc.first_name=c.first_name,cc.last_name=c.last_name,cc.email=c.email
// NOTE: This should only be used with valid columns.
func BuildColumnsUpdateFragment(columns []columns.Column, dialect Dialect) string {
	var cols []string
	for _, column := range columns {
		colName := dialect.QuoteIdentifier(column.Name())
		if column.ToastColumn {
			var colValue string
			if column.KindDetails == typing.Struct {
				colValue = dialect.BuildProcessToastStructColExpression(colName)
			} else {
				colValue = dialect.BuildProcessToastColExpression(colName)
			}
			cols = append(cols, fmt.Sprintf("%s= %s", colName, colValue))
		} else {
			// This is to make it look like: objCol = cc.objCol
			cols = append(cols, fmt.Sprintf("%s=cc.%s", colName, colName))
		}
	}

	return strings.Join(cols, ",")
}