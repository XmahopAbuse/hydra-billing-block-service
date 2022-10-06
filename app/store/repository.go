package store

import (
	"database/sql"
	"hydra-blocking/external/hydra"
	"time"
)

type LocalRepository struct {
	store *Store
}

func InitLocalRepository(store *Store) *LocalRepository {
	return &LocalRepository{store: store}
}

func (r *LocalRepository) CreateBlock(customer *hydra.HydraCustomer) error {
	timeNow := time.Now()
	query := "INSERT INTO blocks(customer_id, customer_code, customer_login, start_date) values ($1, $2, $3, $4)"
	_, err := r.store.db.Exec(query, customer.CustomerId, customer.CustomerCode, customer.CustomerLogin, timeNow)
	if err != nil {
		return err
	}
	return nil
}

func (r *LocalRepository) GetLatestBlock(customer *hydra.HydraCustomer) (*Block, error) {
	var block Block
	query := "SELECT id, customer_id, customer_code, customer_login, start_date, end_date FROM blocks where customer_id = $1 ORDER BY id DESC LIMIT 1"
	err := r.store.db.QueryRow(query, customer.CustomerId).Scan(&block.id, &block.CustomerId, &block.CustomerCode, &block.CustomerLogin, &block.StartDate, &block.EndDate)
	if err != nil {
		if err == sql.ErrNoRows {
			block.CustomerId = ""
			return &block, nil
		}
		return nil, err
	}
	return &block, nil
}

func (r *LocalRepository) UpdateLatestBlock(customer *hydra.HydraCustomer) error {
	block, _ := r.GetLatestBlock(customer)

	timeNow := time.Now()

	query := "UPDATE blocks SET end_date = $1 WHERE  id = $2"

	_, err := r.store.db.Query(query, timeNow, block.id)
	if err != nil {
		return err
	}

	return nil
}
