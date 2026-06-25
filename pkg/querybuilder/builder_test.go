package querybuilder

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fivemanage/lite/api"
)

func TestBuildRejectsUnsafeField(t *testing.T) {
	value := "x"
	_, _, err := New().Filter(api.DatasetFilter{
		Field:    "Attributes['x']) OR 1=1 --",
		Operator: "==",
		Value:    &value,
	}).WithDateRange(time.Now().Add(-time.Hour), time.Now()).Build("org", "dataset")

	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter, got %v", err)
	}
}

func TestBuildParameterizesFilterValues(t *testing.T) {
	value := "x') OR 1=1 --"
	query, args, err := New().Filter(api.DatasetFilter{
		Field:    "message",
		Operator: "contains",
		Value:    &value,
	}).WithDateRange(time.Now().Add(-time.Hour), time.Now()).Build("org", "dataset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(query, value) {
		t.Fatalf("query contains raw filter value: %s", query)
	}
	if len(args) == 0 {
		t.Fatalf("expected query args")
	}
}
