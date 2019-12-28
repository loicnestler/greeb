package common

import (
	"database/sql"
)

type Address struct {
	Address  string         `db:"address"`
	Analyzed bool           `db:"analyzed"`
	Email    sql.NullString `db:"email"`
}
