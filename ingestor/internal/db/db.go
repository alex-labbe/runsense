package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alex-labbe/runsense/ingestor/internal/config"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
A shared *pgxpool.Pool for the entire process
Pool is concurrency safe
*/
type Store struct {
	pool      *pgxpool.Pool
	rawTable  string
	featTable string
}

/*
A validated window ready for insertion to db
Only to be constructed when validated
*/
type RawWindow struct {
	DeviceID  string
	TsStart   time.Time
	FsHz      int
	DurationS float32
	Seq       int64
	Label     *string // empty if nil
	Payload   json.RawMessage
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
On success: return &Store{pool: pool, rawTable: cfg.RawTable, featTable: cfg.FeatTable}.
*/
func New(ctx context.Context, cfg config.Config) (*Store, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDB,
	)
	// Optionally tweaks pool config (if you parse into pgxpool.Config instead). - would likely be better to do this

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// immediately ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	// on success - keep pool open; caller must call Close()
	return &Store{pool: pool, rawTable: cfg.RawTable, featTable: cfg.FeatTable}, nil
}

func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

/*
cheap health check for /health. return nil if DB is reachable, non-nil if not.
used from HTTP handler for health check
*/
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) InsertRawWindow(ctx context.Context, w RawWindow) (int64, error) {
	// init check
	if s == nil || s.pool == nil {
		return 0, errors.New("db store not initialized")
	}

	// build query
	query := fmt.Sprintf(`
	INSERT INTO %s (
		device_id, ts_start, fs_hz, duration_s, seq, label, payload
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`, s.rawTable)

	var id int64
	err := s.pool.
		QueryRow(ctx, query,
			w.DeviceID,
			w.TsStart,
			w.FsHz,
			w.DurationS,
			w.Seq,
			w.Label,
			w.Payload,
		).
		Scan(&id)

	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDuplicateWindow
		}
		return 0, fmt.Errorf("insert raw window: %w", err)
	}

	return id, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
