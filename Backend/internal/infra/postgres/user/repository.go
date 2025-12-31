package user

import (
	"context"
	"database/sql"
	"errors"
	"fpart/internal/domain/user"
	"fpart/internal/pkg/errs"
	"strings"
	"time"
)

type UserRepository struct {
	pool *sql.DB
}

func NewUserRepository(pool *sql.DB) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) AddUser(ctx context.Context, username, email, password string) error {
	tx, err := r.pool.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return &errs.InternalError{Domain: err.Error()}
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			panic(err)
		}
	}()

	var (
		exec_err error
	)

	// insert call with retry if error != nil
	for i := 0; i < 3; i++ {
		if ctx.Err() != nil {
			return &errs.InternalError{Domain: ctx.Err().Error()}
		}

		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		_, exec_err = tx.ExecContext(reqCtx, `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`, username, email, password)
		cancel()
		if exec_err != nil {
			if errors.Is(exec_err, context.DeadlineExceeded) {
				time.Sleep(1 * time.Second)
				continue
			}
			if strings.Contains(exec_err.Error(), "duplicate key value violates unique constraint") ||
				strings.Contains(exec_err.Error(), "insert or update on table \"users\" violates foreign key constraint") {
				return &errs.DuplicateError{}
			}
			return &errs.InternalError{Domain: exec_err.Error()}
		} else {
			break
		}
	}

	if err := tx.Commit(); err != nil {
		return &errs.InternalError{Domain: err.Error()}
	}

	return nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	tx, err := r.pool.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, &errs.InternalError{Domain: err.Error()}
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			panic(err.Error())
		}
	}()

	var read_err error
	var user user.User

	for i := 0; i < 3; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if read_err = tx.QueryRow(`SELECT username, password FROM users`).Scan(&user.Name, &user.Password); read_err != nil {
			if errors.Is(read_err, sql.ErrNoRows) {
				return nil, &errs.NotFoundError{Domain: "user"}
			}
			continue
		}

		break
	}

	if read_err != nil {
		return nil, &errs.InternalError{Domain: "user"}
	}

	return &user, nil
}
