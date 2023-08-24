package kvmigrator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSortMigrationsByID(t *testing.T) {
	data := dataTestSortMigrationsByID()

	for i, data := range data {
		t.Run(fmt.Sprintf("dataset %d", i), func(t *testing.T) {
			input := data.input
			sorted := SortMigrationsByID(input)
			assert.NotSame(t, input, sorted, "sorted slice is the same as input slice")
			assert.Equal(t, len(input), len(sorted), "sorted slice has unexpected length")

			for id, index := range data.expected {
				assert.Equal(t, id, sorted[index].ID, "unexpected migration at index %d", index)
			}
		})
	}
}

type datasetTestSortMigrationsByID struct {
	input    []*RedisMigration
	expected map[string]int
}

func dataTestSortMigrationsByID() []*datasetTestSortMigrationsByID {
	var data []*datasetTestSortMigrationsByID

	data = append(data, &datasetTestSortMigrationsByID{
		input: []*RedisMigration{
			{ID: "05-z"},
			{ID: "01-c"},
			{ID: "01-m"},
			{ID: "05-a"},
			{ID: "03-a"},
		},
		expected: map[string]int{
			"01-c": 0,
			"01-m": 1,
			"03-a": 2,
			"05-a": 3,
			"05-z": 4,
		},
	})

	data = append(data, &datasetTestSortMigrationsByID{
		input: []*RedisMigration{
			{ID: "001-a"},
			{ID: "002-a"},
			{ID: "01-a"},
		},
		expected: map[string]int{
			"001-a": 0,
			"002-a": 1,
			"01-a":  2,
		},
	})

	data = append(data, &datasetTestSortMigrationsByID{
		input: []*RedisMigration{
			{ID: "001-a"},
			{ID: "002-a"},
			{ID: "0001-a"},
		},
		expected: map[string]int{
			"0001-a": 0,
			"001-a":  1,
			"002-a":  2,
		},
	})

	data = append(data, &datasetTestSortMigrationsByID{
		input: []*RedisMigration{
			{ID: "002-b"},
			{ID: "001-a-1"},
			{ID: "002-a"},
			{ID: "001-a-2"},
		},
		expected: map[string]int{
			"001-a-1": 0,
			"001-a-2": 1,
			"002-a":   2,
			"002-b":   3,
		},
	})

	return data
}
