package db

import (
	"github.com/alex-labbe/runsense/ingestor/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"context"
)

/*
	A shared *pgxpool.Pool for the entire process
	Pool is concurrency safe
*/
type Store struct {
	pool *pgxpool.Pool
	rawTable string
	featTable string
}

/*
	A validated window ready for insertion to db
	Only to be constructed when validated
*/
type Window struct { 
	DeviceID string
	TsStart time.Time
	FsHz int
	DurationS float32
	Seq int64
	Label *string // empty if nil
	Payload json.RawMessage
}

// To be returned if InsertRawWindow when pg reports uniqueness violation on device_id + sequence

/*
Must:

Be returned by InsertRawWindow when Postgres reports a unique violation on (device_id, seq).

Be detectable by errors.Is(err, db.ErrDuplicateWindow) in your caller.

You’ll detect this via pgx’s *pgconn.PgError and code 23505.
*/
var ErrDuplicateWindow = errors.New("duplicate raw window")



/*
What it does (pgxpool version):

Builds a connection string, e.g.:

connStr := fmt.Sprintf(
    "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
    cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDatabase,
)


Creates a pgx pool:

pool, err := pgxpool.New(ctx, connStr)



Pings the DB once (pool.Ping(ctx)).

If anything fails: return an error and do not keep the pool.

On success: return &Store{pool: pool, rawTable: cfg.RawTable, featTable: cfg.FeatTable}.

Must:

Fail fast at startup if DB is unreachable or auth is wrong.

Not start any long-running goroutines (just create the pool and ping).
*/
func New(ctx context.Context, cfg config.Config) (*Store, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDatabase,
	)

	// Optionally tweaks pool config (if you parse into pgxpool.Config instead). - would likely be better to do this

	pool, err := pgxpool.New(ctx, connStr)

	// if err i should probably do something

	// immediately ping i think

}


