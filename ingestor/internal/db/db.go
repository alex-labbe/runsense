package db

type DB struct {
	conn *pgx.Conn
}