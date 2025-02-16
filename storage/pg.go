package storage

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/coin"
	"avito-intern/internal/common"
	"avito-intern/internal/merch"
	"avito-intern/pkg/db"
	"context"
	"database/sql"
	"fmt"
	"time"

	"errors"

	"github.com/jackc/pgx/v5"
)

type pgUser struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	Password     string    `db:"password"`
	CoinsBalance int       `db:"coin_balance"`
	CreatedAt    time.Time `db:"created_at"`
}

// PgRepository is a repository for PostgreSQL.
type PgRepository struct {
	db *db.Database
}

// NewRepo creates a new PgUserRepo instance.
func NewRepo(database *db.Database) *PgRepository {
	return &PgRepository{
		db: database,
	}
}

func mapUser(user *pgUser) *auth.User {
	return &auth.User{
		ID:          auth.UserID(user.ID),
		Username:    user.Username,
		Password:    user.Password,
		CoinBalance: user.CoinsBalance,
		CreatedAt:   user.CreatedAt,
	}
}

// CreateUser creates a new user and returns it.
// It assumes a table "users" with columns id, username, and password.
func (r *PgRepository) CreateUser(ctx context.Context, username, password string, coins int) (*auth.User, error) {
	query := `INSERT INTO users (username, password, coin_balance) VALUES ($1, $2, $3) RETURNING id, username, password, coin_balance`
	var user pgUser
	if err := r.db.Get(ctx, &user, query, username, password, coins); err != nil {
		// if no row is returned, consider it as not found.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return mapUser(&user), nil
}

// GetUserByUsername returns the user matching the specified username.
func (r *PgRepository) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `SELECT id, username, password, coin_balance FROM users WHERE username = $1`
	var user pgUser
	if err := r.db.Get(ctx, &user, query, username); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return mapUser(&user), nil
}

// GetUserByID returns the user with the given ID.
func (r *PgRepository) GetUserByID(ctx context.Context, userID auth.UserID) (*auth.User, error) {
	query := `SELECT id, username, password, coin_balance FROM users WHERE id = $1`
	var user pgUser
	if err := r.db.Get(ctx, &user, query, int64(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return mapUser(&user), nil
}

func (r *PgRepository) SaveTransaction(ctx context.Context, t *coin.Transaction) (*coin.Transaction, error) {
	if t == nil {
		return nil, errors.New("invalid transaction")
	}
	err := r.db.RunInTransaction(ctx, func(ctx context.Context) error {
		var currentBalance int
		txErr := r.db.Get(ctx, &currentBalance, `
            SELECT coin_balance 
            FROM users 
            WHERE id = $1 
            FOR UPDATE`, t.FromUser.ID)
		if txErr != nil {
			return fmt.Errorf("failed to get user balance: %w", txErr)
		}

		if currentBalance < t.Amount {
			return coin.ErrNotEnoughCoins
		}
		result, err := r.db.Exec(ctx, `
            UPDATE users 
            SET coin_balance = coin_balance - $2
            WHERE id = $1`, t.FromUser.ID, t.Amount)
		if err != nil {
			return err
		}
		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			return errors.New("failed to update sender balance")
		}
		var txID int64
		if t.ToUser != nil {
			// Lock recipient row
			result, err = r.db.Exec(ctx, `
                UPDATE users 
                SET coin_balance = coin_balance + $2
                WHERE id = $1`, t.ToUser.ID, t.Amount)
			if err != nil {
				return err
			}
			if result.RowsAffected() == 0 {
				return errors.New("failed to update recipient balance")
			}
		} else {
			txErr = r.db.Get(ctx, &txID, `
	INSERT INTO transactions (fk_from_user, amount, type)
	VALUES ($1, $2, $3) RETURNING id;
	`, t.FromUser.ID, t.Amount, t.Type)
		}
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

type pgTransaction struct {
	ID       int64          `db:"id"`
	Amount   int            `db:"amount"`
	FromUser string         `db:"user_from_username"`
	ToUser   sql.NullString `db:"user_to_username"`
}

func (r *PgRepository) GetIncomingTransfers(ctx context.Context, userID auth.UserID) ([]*coin.Transaction, error) {
	query := `
select
    t.id as id,
    t.amount as amount,
    f.username as user_from_username,
    u.username as user_to_username
from transactions t
join users f on f.id = t.fk_from_user
left join users u on u.id = t.fk_to_user
where t.fk_to_user = $1;
`
	var res []pgTransaction
	err := r.db.Select(ctx, &res, query, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("GetIncomingTransfers: %v", err)
	}
	result := make([]*coin.Transaction, len(res))
	for id, row := range res {
		result[id] = &coin.Transaction{
			ID: coin.TransactionID(row.ID),
			FromUser: &auth.User{
				Username: row.FromUser,
			},

			Amount: row.Amount,
		}
		if row.ToUser.Valid {
			result[id].ToUser = &auth.User{
				Username: row.ToUser.String,
			}
		}
	}
	return result, nil
}

func (r *PgRepository) GetOutgoingTransfers(ctx context.Context, userID auth.UserID) ([]*coin.Transaction, error) {
	query := `
select
    t.id as id,
    t.amount as amount,
    f.username as user_from_username,
    u.username as user_to_username
from transactions t
join users f on f.id = t.fk_from_user
left join users u on u.id = t.fk_to_user
where t.fk_from_user = $1 and t.type = $2;
`
	var res []pgTransaction
	err := r.db.Select(ctx, &res, query, userID, coin.Transfer)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("GetOutgoingTransfers: %v", err)
	}
	result := make([]*coin.Transaction, len(res))
	for id, row := range res {
		var toUser *auth.User
		if row.ToUser.Valid {
			toUser = &auth.User{
				Username: row.ToUser.String,
			}
		}
		result[id] = &coin.Transaction{
			ID: coin.TransactionID(row.ID),
			FromUser: &auth.User{
				Username: row.FromUser,
			},
			ToUser: toUser,
			Amount: row.Amount,
		}
	}
	return result, nil
}

type pgMerch struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Price int    `db:"price"`
}

type pgPurchase struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"fk_user"`
	MerchID     int64     `db:"fk_merch"`
	MerchName   string    `db:"merch_name"` // Add this field
	Quantity    int       `db:"quantity"`
	PurchasedAt time.Time `db:"purchased_at"`
}

func mapMerch(m *pgMerch) *merch.Merch {
	return &merch.Merch{
		ID:    m.ID,
		Name:  m.Name,
		Price: m.Price,
	}
}

func mapPurchase(p *pgPurchase) *merch.Purchase {
	return &merch.Purchase{
		ID:          int(p.ID),
		UserID:      auth.UserID(p.UserID),
		MerchID:     p.MerchID,
		MerchName:   p.MerchName, // Map the new field
		Quantity:    p.Quantity,
		PurchasedAt: p.PurchasedAt,
	}
}

// GetMerchByID returns the merch with the given ID.
func (r *PgRepository) GetMerchByID(ctx context.Context, merchName string) (*merch.Merch, error) {
	query := `SELECT id, name, price FROM merch WHERE name = $1`
	var m pgMerch
	if err := r.db.Get(ctx, &m, query, merchName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, merch.ErrMerchNotFound
		}
		return nil, fmt.Errorf("failed to get merch by id: %w", err)
	}
	return mapMerch(&m), nil
}

// SavePurchase saves a purchase record.
func (r *PgRepository) SavePurchase(ctx context.Context, p *merch.Purchase) error {
	query := `INSERT INTO purchases (fk_user, fk_merch, quantity, purchased_at) VALUES ($1, $2, $3, $4) RETURNING id`
	var id int64
	if err := r.db.Get(ctx, &id, query, p.UserID, p.MerchID, p.Quantity, p.PurchasedAt); err != nil {
		return fmt.Errorf("failed to save purchase: %w", err)
	}
	p.ID = int(id)
	return nil
}

// ListPurchasesByUserID lists all purchases made by a user.
func (r *PgRepository) ListPurchasesByUserID(ctx context.Context, userID auth.UserID) ([]*merch.Purchase, error) {
	query := `
        SELECT
            p.id,
            p.fk_user,
            p.fk_merch,
            m.name as merch_name,
            p.quantity,
            p.purchased_at
        FROM purchases p
        JOIN merch m ON m.id = p.fk_merch
        WHERE p.fk_user = $1`
	var purchases []pgPurchase
	if err := r.db.Select(ctx, &purchases, query, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, merch.ErrPurchasesNotFound
		}
		return nil, fmt.Errorf("failed to list purchases by user id: %w", err)
	}
	result := make([]*merch.Purchase, len(purchases))
	for i, p := range purchases {
		result[i] = mapPurchase(&p)
	}
	return result, nil
}
