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

type RedisMigrator struct {
	client          *redis.Client
	migrations      []*RedisMigration
	migrationLogKey string
}

func NewRedisMigrator(client *redis.Client) *RedisMigrator {
	migrator := &RedisMigrator{
		client:     client,
		migrations: make([]*RedisMigration, 0),
	}

	migrator.SetKeyPrefix("redis-migrator:")

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
		id := migration.ID

		if applied, err := red.SIsMember(ctx, migrator.migrationLogKey, id).Result(); err != nil {
			return fmt.Errorf("checking migration log for migration %s failed: %w", id, err)
		} else if applied {
			// skip already applied migration
			continue
		}

		// apply
		if err := migrator.apply(ctx, migration); err != nil {
			return fmt.Errorf("applying migration %s failed: %w", id, err)
		}

		// add to migration log
		if err := red.SAdd(ctx, migrator.migrationLogKey, id).Err(); err != nil {
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
