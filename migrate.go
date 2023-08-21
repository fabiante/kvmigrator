package kvmigrator

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type RedisUpFunc = func(ctx context.Context, client *redis.Client) error

type RedisMigration struct {
	ID string
	Up RedisUpFunc
}

func NewRedisMigration(ID string, up RedisUpFunc) *RedisMigration {
	return &RedisMigration{ID: ID, Up: up}
}

// RedisMigrator runs migrations (RedisMigration) for a given redis.Client.
type RedisMigrator struct {
	client          *redis.Client
	migrations      []*RedisMigration
	migrationLogKey string
}

// NewRedisMigrator creates a new migrator using the given client.
//
// The given prefix is used for internal keys which the migrator maintains to manage migrations.
func NewRedisMigrator(client *redis.Client, prefix string) *RedisMigrator {
	migrator := &RedisMigrator{
		client:     client,
		migrations: make([]*RedisMigration, 0),
	}

	migrator.SetKeyPrefix(prefix)

	return migrator
}

// SetKeyPrefix sets the prefix used on all keys used to manage migrations.
func (migrator *RedisMigrator) SetKeyPrefix(prefix string) *RedisMigrator {
	migrator.migrationLogKey = prefix + "migration-log"
	return migrator
}

func (migrator *RedisMigrator) AddMigrations(m ...*RedisMigration) *RedisMigrator {
	migrator.migrations = append(migrator.migrations, m...)
	return migrator
}

func (migrator *RedisMigrator) Migrate(ctx context.Context) error {
	red := migrator.client

	// execute migrations one after another. skips already applied migrations
	for _, migration := range migrator.migrations {
		// check if migrations should be cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("migrations cancelled")
		default:
			// continue with next migration
		}

		id := migration.ID

		// migrations should not be cancelled mid-process. therefore a child context
		// without cancellation is passed.
		migrationCtx := context.WithoutCancel(ctx)

		if applied, err := red.SIsMember(migrationCtx, migrator.migrationLogKey, id).Result(); err != nil {
			return fmt.Errorf("checking migration log for migration %s failed: %w", id, err)
		} else if applied {
			// skip already applied migration
			continue
		}

		// apply
		if err := migrator.apply(migrationCtx, migration); err != nil {
			return fmt.Errorf("applying migration %s failed: %w", id, err)
		}

		// add to migration log
		if err := red.SAdd(migrationCtx, migrator.migrationLogKey, id).Err(); err != nil {
			return fmt.Errorf("migration %s applied, but adding to migration log failed: %w", id, err)
		}
	}

	return nil
}

func (migrator *RedisMigrator) apply(ctx context.Context, migration *RedisMigration) (retErr error) {
	// recover from panics
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("migration paniced: %v", err)
		}
	}()

	if err := migration.Up(ctx, migrator.client); err != nil {
		return fmt.Errorf("migration up failed: %w", err)
	}

	return
}
