package kvmigrator

import (
	"slices"
	"strings"
)

// SortMigrationsByID sorts the given migrations based on comparing their IDs.
//
// strings.Compare is used as comparator. Look at the tests of this function if you
// want to see examples of how this is sorted.
//
// The input slice is not modified in-place but instead a new, sorted slice is returned.
//
// Once this module reaches v1 its sorting algorithm should not be changed.
func SortMigrationsByID(migrations []*RedisMigration) []*RedisMigration {
	return SortRedisMigrations(migrations, func(a, b *RedisMigration) int {
		return strings.Compare(a.ID, b.ID)
	})
}

// SortRedisMigrations is a helper function to sort the given migrations. You can
// supply your own comparator function to be used for sorting.
//
// The input slice is not modified in-place but instead a new, sorted slice is returned.
func SortRedisMigrations(migrations []*RedisMigration, cmp func(a, b *RedisMigration) int) []*RedisMigration {
	// clone slice
	migrations = slices.Clone(migrations)

	// sort slice
	slices.SortFunc(migrations, cmp)

	return migrations
}
