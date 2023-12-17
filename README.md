# pgsafemigrate - PostgreSQL Schema & Data Migration Linter

`pgsafemigrate` is a command-line tool and Go library
that performs a series of checks
on PostgreSQL schema or data migrations SQL statements.

Its design supports the [sql-migrate](https://github.com/rubenv/sql-migrate)
migration file format and is automatically recognized. The tool is thus aware of:
1. The migration direction (up/down).
2. Whether statements are wrapped or not in a transaction.

## Features

### Migration Direction Context

`sql-migrate` is a popular tool for handling database migrations.
The migration file format requires & defines migration directives in special
SQL comments:

```sql
-- +migrate Up
-- Forward migration statements in this section

-- +migrate Down
-- Rollback migration statements included in this section
```

`pgsafemigrate` is aware of the migration direction and passes this condition
to each rule as execution context. This means that
it is able to ignore certain rules that would only apply to forward migrations or vice versa.
For example, a rule that warns against columns being dropped can potentially
be ignored when they're part of a down migration.

### No-Lint Annotations

There will always be exceptions to the rules. For example, engineers may be
aware that a table is empty while the feature is under development
and receives no traffic at the time. To accommodate these cases, `pgsafemigrate`
uses a special annotation format defined in a SQL comment:

```sql
-- +migrate Up
-- pgsafemigrate:nolint:high-availability-avoid-table-rename
ALTER TABLE "movies" RENAME TO "films";
```

> :warning: No-lint annotations apply to **all** statements in the same migration direction (`sql-migrate` format) or the same migration file.

### Transactions & Idempotency

If a migration consists of multiple statements, and the migration fails
mid-step it is likely that it cannot be resumed in an idempotent manner.
Fixing the database state or modifying the migration can be a risky
and stressful operation.

> `sql-migrate` by default wraps all migration statements in a transaction.

### Nested Transaction Detection

PostgreSQL supports marking different sections of a transaction with [SAVEPOINT](https://www.postgresql.org/docs/current/sql-savepoint.html),
instead of supporting nested transactions.

### Transaction-Incompatible Statements

Certain statements cannot be executed as part of a transaction.
A frequent operation that falls into this category is creating an index concurrently.
`pgsafemigrate` assumes all statements are wrapped in a transaction,
unless the `sql-migrate` `notransaction` command is defined.

## Contributing

### Adding New Rules

A Rule must implement the following interface:

```go
type Rule interface {
    // Documentation returns a string containing more in-depth details and insights on the rule logic,
    // along with any guidelines about how to fix or handle this operation optimally.
    Documentation() string
    // Alias returns the unique identifier of the rule. The alias is prefixed with a shared category prefix,
    // followed by a code that briefly explains the rule scope.
    Alias() string
    // Process receives a parsed SQL statement that is part of the migration,
    // along with the entire set of migration statements and a flag denoting that
    // the statement is executed within an active transaction or not.
    // Returns true if the rule matches and a warning must be produced for this statement.
    Process(node *pg_query.Node, allStatements []*pg_query.Node, inTransaction bool) bool
}
```

[`pg_query` nodes](https://github.com/pganalyze/pg_query_go) provide access to the full range of PostgreSQL syntax.
See existing rules for examples. You can start by writing a test case with a statement sample that you want to test and then
inspect the Parse Tree to find out the node properties that need to be accessed and checked accordingly.
