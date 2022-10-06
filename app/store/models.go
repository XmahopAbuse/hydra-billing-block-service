package store

import (
	"database/sql"
	"time"
)

type Block struct {
	id                                      int
	CustomerId, CustomerCode, CustomerLogin string
	StartDate                               time.Time
	EndDate                                 sql.NullTime
}
