package snowflake

import (
	"log/slog"

	"github.com/artie-labs/transfer/clients/shared"
	"github.com/artie-labs/transfer/lib/config/constants"
	"github.com/artie-labs/transfer/lib/destination/types"
	"github.com/artie-labs/transfer/lib/logger"
	"github.com/artie-labs/transfer/lib/optimization"
	"github.com/artie-labs/transfer/lib/sql"
	"github.com/artie-labs/transfer/lib/typing"
	"github.com/artie-labs/transfer/lib/typing/columns"
)

func (s *Store) Append(tableData *optimization.TableData, _ bool) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			if IsAuthExpiredError(err) {
				slog.Warn("Authentication has expired, will reload the Snowflake store and retry appending", slog.Any("err", err))
				if connErr := s.reestablishConnection(); connErr != nil {
					// TODO: Remove this panic and return an error instead. Ensure the callers of [Append] handle this properly.
					logger.Panic("Failed to reestablish connection", slog.Any("err", connErr))
				}
			} else {
				break
			}
		}

		// TODO: For history mode - in the future, we could also have a separate stage name for history mode so we can enable parallel processing.
		err = shared.Append(s, tableData, types.AdditionalSettings{
			AdditionalCopyClause: `FILE_FORMAT = (TYPE = 'csv' FIELD_DELIMITER= '\t' FIELD_OPTIONALLY_ENCLOSED_BY='"' NULL_IF='\\N' EMPTY_FIELD_AS_NULL=FALSE) PURGE = TRUE`,
		})
	}

	return err
}

func (s *Store) additionalEqualityStrings(tableData *optimization.TableData) []string {
	cols := make([]columns.Column, len(tableData.TopicConfig().AdditionalMergePredicates))
	for i, additionalMergePredicate := range tableData.TopicConfig().AdditionalMergePredicates {
		cols[i] = columns.NewColumn(additionalMergePredicate.PartitionField, typing.Invalid)
	}
	return sql.BuildColumnComparisons(cols, constants.TargetAlias, constants.StagingAlias, sql.Equal, s.Dialect())
}

func (s *Store) Merge(tableData *optimization.TableData) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			if IsAuthExpiredError(err) {
				slog.Warn("Authentication has expired, will reload the Snowflake store and retry merging", slog.Any("err", err))
				if connErr := s.reestablishConnection(); connErr != nil {
					// TODO: Remove this panic and return an error instead. Ensure the callers of [Merge] handle this properly.
					logger.Panic("Failed to reestablish connection", slog.Any("err", connErr))
				}
			} else {
				break
			}
		}

		err = shared.Merge(s, tableData, types.MergeOpts{
			AdditionalEqualityStrings: s.additionalEqualityStrings(tableData),
		})
	}
	return err
}
