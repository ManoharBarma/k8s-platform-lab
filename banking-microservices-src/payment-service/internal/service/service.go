// Package service implements payment-service use cases: paying a beneficiary by
// debiting the payer account and crediting a system clearing account.
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/outbox"
	"banking-platform/services/payment-service/internal/domain"
	"banking-platform/services/payment-service/internal/ledgerclient"
	"banking-platform/services/payment-service/internal/repository"
)

// EventTopic is the Kafka topic payment events are published to.
const EventTopic = "banking.payments"

var clearingNamespace = uuid.MustParse("00000000-0000-0000-0000-0000000000ee")

func clearingAccount(currency string) string {
	return uuid.NewSHA1(clearingNamespace, []byte("system-clearing-"+currency)).String()
}

// Ledger is the subset of the ledger payment-service depends on.
type Ledger interface {
	OpenAccount(ctx context.Context, accountID, currency string) error
	Post(ctx context.Context, txID, idemKey, currency, reference string, entries []ledgerclient.Entry) (string, bool, error)
	Balance(ctx context.Context, accountID string) (balance, currency string, err error)
}

// Service holds payment-service dependencies.
type Service struct {
	repo       repository.Repository
	ledger     Ledger
	source     string
	emitEvents bool
	log        *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, ledger Ledger, source string, emitEvents bool, log *zap.Logger) *Service {
	return &Service{repo: repo, ledger: ledger, source: source, emitEvents: emitEvents, log: log}
}

// PayInput is the data required to make a payment.
type PayInput struct {
	PayerAccountID string
	BeneficiaryID  string
	Amount         string
	Currency       string
	Reference      string
	IdempotencyKey string
}

// Pay executes an outbound payment.
func (s *Service) Pay(ctx context.Context, in PayInput) (domain.Payment, error) {
	amount, err := decimal.NewFromString(in.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return domain.Payment{}, apierror.ErrBadRequest("amount must be a positive number")
	}
	if in.PayerAccountID == "" || in.BeneficiaryID == "" {
		return domain.Payment{}, apierror.ErrBadRequest("payer_account_id and beneficiary_id are required")
	}

	if in.IdempotencyKey != "" {
		if existing, err := s.repo.GetByIdempotencyKey(ctx, in.IdempotencyKey); err == nil {
			return existing, nil
		}
	}

	// Overdraft check.
	balStr, _, err := s.ledger.Balance(ctx, in.PayerAccountID)
	if err != nil {
		return domain.Payment{}, ledgerErr()
	}
	bal, _ := decimal.NewFromString(balStr)
	if bal.LessThan(amount) {
		return domain.Payment{}, apierror.New(422, "insufficient_funds", "insufficient funds")
	}

	clearing := clearingAccount(in.Currency)
	if err := s.ledger.OpenAccount(ctx, clearing, in.Currency); err != nil {
		return domain.Payment{}, ledgerErr()
	}

	p := domain.Payment{
		ID:             uuid.NewString(),
		PayerAccountID: in.PayerAccountID,
		BeneficiaryID:  in.BeneficiaryID,
		Amount:         amount,
		Currency:       in.Currency,
		Status:         domain.StatusCompleted,
		Reference:      in.Reference,
		IdempotencyKey: in.IdempotencyKey,
		CreatedAt:      time.Now().UTC(),
	}

	entries := []ledgerclient.Entry{
		{AccountID: in.PayerAccountID, Direction: "DEBIT", Amount: amount.String()},
		{AccountID: clearing, Direction: "CREDIT", Amount: amount.String()},
	}
	if _, _, err := s.ledger.Post(ctx, p.ID, keyOr(in.IdempotencyKey, p.ID), in.Currency, "payment:"+p.ID, entries); err != nil {
		return domain.Payment{}, ledgerErr()
	}

	var event *outbox.Event
	if s.emitEvents {
		payload, _ := json.Marshal(p)
		event = &outbox.Event{ID: uuid.NewString(), Source: s.source, Topic: EventTopic, Key: p.PayerAccountID, Payload: payload}
	}
	if err := s.repo.Create(ctx, p, event); err != nil {
		return domain.Payment{}, err
	}
	return p, nil
}

// Get returns a payment by ID.
func (s *Service) Get(ctx context.Context, id string) (domain.Payment, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByPayer returns payments made from an account.
func (s *Service) ListByPayer(ctx context.Context, payerAccountID string) ([]domain.Payment, error) {
	return s.repo.ListByPayer(ctx, payerAccountID)
}

func keyOr(k, fallback string) string {
	if k != "" {
		return k
	}
	return fallback
}

func ledgerErr() error {
	return apierror.New(502, "ledger_unavailable", "ledger service unavailable")
}
