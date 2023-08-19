# KV Migrator

This project aims to develop a framework for migrating data stored inside key-value databases.

It is inspired by other migration frameworks in the Go ecosystem which often are aimed
towards relational databases.

`kvmigrator` tries to fill this gap by explicitly focussing on key-value databases. For now development
focuses on supporting primarily [Redis](https://redis.com/) but theoretically supporting other kv databases
should be possible in the future.

**Project Status:** As long as this readme does not explicitly state any compatibility promises,
please don't expect any support activity. As of now consider this project is an experiment. Still,
contributions and feedback are welcome.

## Contributing

### Testing

Currently, tests require a running redis instance available via `127.0.0.1:6379`. These tests modify
data withing the database so be sure to not accidentally use the wrong database.
