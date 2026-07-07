package querybuilder

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fivemanage/lite/api"
	"github.com/huandu/go-sqlbuilder"
)

var (
	ErrInvalidFilter = errors.New("invalid log filter")
	fieldPattern     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]{0,127}$`)
)

var topLevelFields = map[string]struct{}{
	"Timestamp": {}, "DatasetId": {}, "TraceId": {}, "TeamId": {}, "Body": {}, "RetentionDays": {},
}

type Builder struct {
	query *sqlbuilder.SelectBuilder
	err   error
}

func New() *Builder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("*")
	sb.From("logs")
	return &Builder{query: sb}
}
func (b *Builder) Filter(filter api.DatasetFilter) *Builder {
	cond := b.buildCondition(filter)
	if b.err != nil {
		return b
	}
	if cond != "" {
		b.query.Where(cond)
	}
	return b
}
func (b *Builder) buildCondition(filter api.DatasetFilter) string {
	if filter.Field != "" || filter.Operator != "" || filter.Value != nil {
		field, ok := b.fieldExpr(filter.Field)
		if !ok {
			b.err = ErrInvalidFilter
			return ""
		}
		switch filter.Operator {
		case "exists":
			return fmt.Sprintf("mapContains(Attributes, '%s')", escapeLiteral(filter.Field))
		case "not-exists":
			return fmt.Sprintf("NOT mapContains(Attributes, '%s')", escapeLiteral(filter.Field))
		}
		if filter.Value == nil {
			b.err = ErrInvalidFilter
			return ""
		}
		value := *filter.Value
		switch filter.Operator {
		case "contains":
			return b.query.Like(field, "%"+value+"%")
		case "not-contains":
			return fmt.Sprintf("NOT (%s)", b.query.Like(field, "%"+value+"%"))
		case "starts-with":
			return b.query.Like(field, value+"%")
		case "ends-with":
			return b.query.Like(field, "%"+value)
		case "==":
			return b.query.Equal(field, value)
		case "!=":
			return b.query.NotEqual(field, value)
		case ">":
			return b.query.GreaterThan(field, value)
		case "<":
			return b.query.LessThan(field, value)
		case ">=":
			return b.query.GreaterEqualThan(field, value)
		case "<=":
			return b.query.LessEqualThan(field, value)
		default:
			b.err = ErrInvalidFilter
			return ""
		}
	}
	if len(filter.Children) > 0 {
		var conditions []string
		for _, child := range filter.Children {
			if cond := b.buildCondition(child); cond != "" {
				conditions = append(conditions, cond)
			}
			if b.err != nil {
				return ""
			}
		}
		if len(conditions) > 0 {
			logicalOperator := " AND "
			if strings.ToUpper(filter.Operator) == "OR" {
				logicalOperator = " OR "
			}
			return fmt.Sprintf("(%s)", strings.Join(conditions, logicalOperator))
		}
	}
	return ""
}
func (b *Builder) fieldExpr(field string) (string, bool) {
	if !fieldPattern.MatchString(field) {
		return "", false
	}
	if _, ok := topLevelFields[field]; ok {
		return field, true
	}
	// ClickHouse map keys/identifiers cannot be bound as query parameters, so the
	// field name is interpolated. fieldPattern already forbids quotes; escaping
	// here is defense-in-depth so a future loosening of the pattern cannot open a
	// SQL-injection break-out via a crafted attribute key.
	return fmt.Sprintf("Attributes['%s']", escapeLiteral(field)), true
}

// escapeLiteral escapes a value for safe inclusion inside a single-quoted
// ClickHouse string literal by doubling backslashes and single quotes.
func escapeLiteral(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `'`, `\'`)
}
func (b *Builder) WithDateRange(startTime, endTime time.Time) *Builder {
	b.query.Where(b.query.Between("Timestamp", startTime, endTime))
	return b
}
func (b *Builder) Build(organizationID, datasetID string) (string, []interface{}, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	b.query.Where(b.query.Equal("TeamId", organizationID))
	b.query.Where(b.query.Equal("DatasetId", datasetID))
	query, args := b.query.BuildWithFlavor(sqlbuilder.ClickHouse)
	return query, args, nil
}
