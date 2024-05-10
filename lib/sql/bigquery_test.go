package sql

import (
	"testing"
	"time"

	"github.com/artie-labs/transfer/lib/config/constants"
	"github.com/artie-labs/transfer/lib/typing"
	"github.com/artie-labs/transfer/lib/typing/ext"
	"github.com/stretchr/testify/assert"
)

func TestBigQueryDialect_KindForDataType(t *testing.T) {
	dialect := BigQueryDialect{}

	bqColToExpectedKind := map[string]typing.KindDetails{
		// Number
		"numeric":           typing.EDecimal,
		"numeric(5)":        typing.Integer,
		"numeric(5, 0)":     typing.Integer,
		"numeric(5, 2)":     typing.EDecimal,
		"numeric(8, 6)":     typing.EDecimal,
		"bignumeric(38, 2)": typing.EDecimal,

		// Integer
		"int":     typing.Integer,
		"integer": typing.Integer,
		"inT64":   typing.Integer,
		// String
		"varchar":     typing.String,
		"string":      typing.String,
		"sTriNG":      typing.String,
		"STRING (10)": typing.String,
		// Array
		"array<integer>": typing.Array,
		"array<string>":  typing.Array,
		// Boolean
		"bool":    typing.Boolean,
		"boolean": typing.Boolean,
		// Struct
		"STRUCT<foo STRING>": typing.Struct,
		"record":             typing.Struct,
		"json":               typing.Struct,
		// Datetime
		"datetime":  typing.NewKindDetailsFromTemplate(typing.ETime, ext.DateTimeKindType),
		"timestamp": typing.NewKindDetailsFromTemplate(typing.ETime, ext.DateTimeKindType),
		"time":      typing.NewKindDetailsFromTemplate(typing.ETime, ext.TimeKindType),
		"date":      typing.NewKindDetailsFromTemplate(typing.ETime, ext.DateKindType),
		//Invalid
		"foo":            typing.Invalid,
		"foofoo":         typing.Invalid,
		"":               typing.Invalid,
		"numeric(1,2,3)": typing.Invalid,
	}

	for bqCol, expectedKind := range bqColToExpectedKind {
		kd, err := dialect.KindForDataType(bqCol, "")
		assert.NoError(t, err)
		assert.Equal(t, expectedKind.Kind, kd.Kind, bqCol)
	}

	{
		_, err := dialect.KindForDataType("numeric(5", "")
		assert.ErrorContains(t, err, "missing closing parenthesis")
	}
	{
		kd, err := dialect.KindForDataType("numeric(5, 2)", "")
		assert.NoError(t, err)
		assert.Equal(t, typing.EDecimal.Kind, kd.Kind)
		assert.Equal(t, 5, *kd.ExtendedDecimalDetails.Precision())
		assert.Equal(t, 2, kd.ExtendedDecimalDetails.Scale())
	}
	{
		kd, err := dialect.KindForDataType("bignumeric(5, 2)", "")
		assert.NoError(t, err)
		assert.Equal(t, typing.EDecimal.Kind, kd.Kind)
		assert.Equal(t, 5, *kd.ExtendedDecimalDetails.Precision())
		assert.Equal(t, 2, kd.ExtendedDecimalDetails.Scale())
	}
}

func TestBigQueryDialect_KindForDataType_NoDataLoss(t *testing.T) {
	kindDetails := []typing.KindDetails{
		typing.NewKindDetailsFromTemplate(typing.ETime, ext.DateTimeKindType),
		typing.NewKindDetailsFromTemplate(typing.ETime, ext.TimeKindType),
		typing.NewKindDetailsFromTemplate(typing.ETime, ext.DateKindType),
		typing.String,
		typing.Boolean,
		typing.Struct,
	}

	for _, kindDetail := range kindDetails {
		kd, err := BigQueryDialect{}.KindForDataType(BigQueryDialect{}.DataTypeForKind(kindDetail, false), "")
		assert.NoError(t, err)
		assert.Equal(t, kindDetail, kd)
	}
}

func fromExpiresDateStringToTime(tsString string) (time.Time, error) {
	return time.Parse(bqLayout, tsString)
}

func TestBQExpiresDate(t *testing.T) {
	// We should be able to go back and forth.
	// Note: The format does not have ns precision because we don't need it.
	birthday := time.Date(2022, time.September, 6, 3, 19, 24, 0, time.UTC)
	for i := 0; i < 5; i++ {
		tsString := BQExpiresDate(birthday)
		ts, err := fromExpiresDateStringToTime(tsString)
		assert.NoError(t, err)
		assert.Equal(t, birthday, ts)
	}

	for _, badString := range []string{"foo", "bad_string", " 2022-09-01"} {
		_, err := fromExpiresDateStringToTime(badString)
		assert.ErrorContains(t, err, "cannot parse", badString)
	}
}

func TestBigQueryDialect_BuildAlterColumnQuery(t *testing.T) {
	assert.Equal(t,
		"ALTER TABLE {TABLE} drop COLUMN {SQL_PART}",
		BigQueryDialect{}.BuildAlterColumnQuery("{TABLE}", constants.Delete, "{SQL_PART}"),
	)
}