package kvmigrator

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMigrate(t *testing.T) {
	ctx := context.Background()

	red := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	require.NoError(t, red.Info(ctx).Err())

	migrator := NewRedisMigrator(red)
	migrator.SetKeyPrefix("app:migrations:")
	migrator.AddMigrations(buildMigrations()...)
	require.NotNil(t, migrator)

	require.NoError(t, migrator.Migrate(ctx))
}

func newKey(k string) string {
	return "app:" + k
}

func buildMigrations() []*RedisMigration {
	return []*RedisMigration{
		// The following migration is pretty minimal. It creates 10 random
		// uuids and pushes them into a list.
		//
		// This does not require the usage of a transaction since all elements are
		// added in a single redis command.
		NewRedisMigration("00000001-seed-rand-ids", func(ctx context.Context, client *redis.Client) error {
			var ids []any
			for i := 0; i < 10; i++ {
				ids = append(ids, uuid.New().String())
			}
			return client.LPush(ctx, newKey("id-list"), ids...).Err()
		}),
		// This migration pops 5 of the created uuids from the list.
		//
		// This requires the use of a transaction to be an atomic operation.
		NewRedisMigration("00000002-pop-5-ids", func(ctx context.Context, client *redis.Client) error {
			_, err := client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				for i := 0; i < 5; i++ {
					if err := client.RPop(ctx, newKey("id-list")).Err(); err != nil {
						return err
					}
				}
				return nil
			})
			return err
		}),
		// This migration more what we could expect from a real app:
		// Seed some data which is tightly coupled to the application code.
		//
		// This also makes use of a transaction to ensure all roles are created.
		NewRedisMigration("00000003-setup-roles", func(ctx context.Context, client *redis.Client) error {
			type role struct {
				id   string
				desc string
			}

			roles := []role{
				{
					id:   "admin",
					desc: "Super user access",
				},
				{
					id:   "moderator",
					desc: "General moderation rights",
				},
				{
					id:   "user",
					desc: "Regular user",
				},
			}

			// seed data in a transaction - ensures this is an atomic update
			_, err := client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				for _, lang := range roles {
					key := newKey(fmt.Sprintf("languages:%s:desc", lang.id))
					if err := pipe.Set(ctx, key, lang.desc, 0).Err(); err != nil {
						return err
					}
				}

				return nil
			})

			return err
		}),
	}
}
