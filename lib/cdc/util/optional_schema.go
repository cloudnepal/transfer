package util

import (
	"log/slog"

	"github.com/artie-labs/transfer/lib/debezium"
	"github.com/artie-labs/transfer/lib/typing"
)

func (s *SchemaEventPayload) GetOptionalSchema() map[string]typing.KindDetails {
	fieldsObject := s.Schema.GetSchemaFromLabel(debezium.After)
	if fieldsObject == nil {
		// AFTER schema does not exist.
		return nil
	}

	schema := make(map[string]typing.KindDetails)
	for _, field := range fieldsObject.Fields {
		kd := field.ToKindDetails()
		if kd == typing.Invalid {
			slog.Warn("Skipping field from optional schema b/c we cannot determine the data type", slog.String("field", field.FieldName))
			continue
		}

		schema[field.FieldName] = kd
	}

	return schema
}
