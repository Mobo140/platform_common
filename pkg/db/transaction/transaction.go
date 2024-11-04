package transaction

import (
	"context"

	"github.com/Mobo140/platform_common/pkg/db"
	"github.com/Mobo140/platform_common/pkg/db/pg"
	"github.com/jackc/pgx/v4"

	"github.com/pkg/errors"
)

var _ db.TxManager = (*manager)(nil)

type manager struct {
	db db.Transactor
}

// NewTranssactionManager creates a new tranasaction manager which satisfies the interface db.TxManager.
func NewTransactionManager(db db.Transactor) *manager { //nolint:revive // it's ok
	return &manager{
		db: db,
	}
}

// transaction - is a main function which call handler inside a transaction.
func (m *manager) transaction(ctx context.Context, opts pgx.TxOptions, fn db.Handler) (err error) {
	// if it is a nested transaction skip initialisation of the new transaction and call handler
	tx, ok := ctx.Value(pg.TxKey).(pgx.Tx)
	if ok {
		return fn(ctx)
	}

	// Start a new transaction
	tx, err = m.db.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "can't begin transaction")
	}

	// Put transaction in context
	ctx = pg.MakeContextTx(ctx, tx)

	// Define a function for rollback and commit transaction
	defer func() {
		// Reinstalling after the panic
		if r := recover(); r != nil {
			err = errors.Errorf("panic recovered: %v", r)
		}

		// Rollback transaction if we had an error
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollback)
			}

			return
		}

		// If we havent'got any errors commit transaction
		if nil == err {
			err = tx.Commit(ctx)
			if err != nil {
				err = errors.Wrap(err, "tx commit failed")
			}
		}
	}()

	// Complete code inside transaction
	// If there is an error in the handler return error and rollback the transaction
	// opposite it commit transaction
	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	return err
}

func (m *manager) ReadCommited(ctx context.Context, f db.Handler) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return m.transaction(ctx, txOpts, f)
}
