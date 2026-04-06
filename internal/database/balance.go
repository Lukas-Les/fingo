package database

import (
	"context"

	"github.com/google/uuid"
)

func (q *Queries) GetUserBalanceAsStr(ctx context.Context, userID uuid.UUID) (string, error) {
	const query = `
		SELECT
			SUM(CASE WHEN transaction_type = 'income' THEN amount ELSE 0 END)::numeric -
			SUM(CASE WHEN transaction_type = 'expense' THEN amount ELSE 0 END)::numeric AS balance
		FROM transactions
		WHERE user_id = $1`
	row := q.db.QueryRowContext(ctx, query, userID)

	var balance string
	err := row.Scan(&balance)
	return balance, err
}
