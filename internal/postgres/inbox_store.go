package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/sanLimbu/eda-go/internal/am"
	"github.com/sanLimbu/eda-go/internal/tm"
)

type InboxStore struct {
	tableName string
	db        DB
}

var _ tm.InboxStore = (*InboxStore)(nil)

func NewInboxStore(tableName string, db DB) InboxStore {
	return InboxStore{
		tableName: tableName,
		db:        db,
	}
}

func (s InboxStore) Save(ctx context.Context, msg am.IncomingMessage) error {
	const query = "INSERT INTO  %s (id, NAME, subject, DATA, metadatam, sent_to, received_at) VALUES ($1, $2,$3,$4,$5,$6,$7)"

	medatadata, err := json.Marshal(msg.Metadata())
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, s.table(query), msg.ID(), msg.MessageName(), msg.Subject(), msg.Data(), medatadata, msg.SentAt(), msg.ReceivedAt())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return tm.ErrDuplicateMessage(msg.ID())
			}
		}
	}
	return err
}

func (s InboxStore) table(query string) string {
	return fmt.Sprintf(query, s.tableName)
}
