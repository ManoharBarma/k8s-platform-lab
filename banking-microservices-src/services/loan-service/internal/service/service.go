// Package service implements loan-service use cases: application, approval and
// disbursement (which credits the borrower's account via the ledger).
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/services/loan-service/internal/domain"
	"banking-platform/services/loan-service/internal/ledgerclient"
	"banking-platform/services/loan-service/internal/repository"
)

var fundingNamespace = uuid.MustParse("00000000-0000-0000-0000-0000000000dd")

func fundingAccount(currency string) string {
	return uuid.NewSHA1(fundingNamespace, []byte("system-loan-funding-"+currency)).String()
}

// Ledger is the subset of the ledger loan-service depends on.
type Ledger interface {
	OpenAccount(ctx context.Context, accountID, currency string) error
	Post(ctx context.Context, txID, idemKey, currency, reference string, entries []ledgerclient.Entry) (string, bool, error)
}

// Service holds loan-service dependencies.
type Service struct {
	repo   repository.Repository
	ledger Ledger
	log    *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, ledger Ledger, log *zap.Logger) *Service {
	return &Service{repo: repo, ledger: ledger, log: log}
}

// ApplyInput is a loan application.
type ApplyInput struct {
	BorrowerID   string
	AccountID    string
	Principal    string
	Currency     string
	AnnualRate   float64
	TenureMonths int
}

// Apply records a new loan application in PENDING state.
func (s *Service) Apply(ctx context.Context, in ApplyInput) (domain.Loan, error) {
	principal, err := decimal.NewFromString(in.Principal)
	if err != nil || principal.LessThanOrEqual(decimal.Zero) {
		return domain.Loan{}, apierror.ErrBadRequest("principal must be a positive number")
	}
	if in.TenureMonths <= 0 {
		return domain.Loan{}, apierror.ErrBadRequest("tenure_months must be positive")
	}
	if in.AccountID == "" {
		return domain.Loan{}, apierror.ErrBadRequest("account_id is required")
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	loan := domain.Loan{
		ID:           uuid.NewString(),
		BorrowerID:   in.BorrowerID,
		AccountID:    in.AccountID,
		Principal:    principal,
		Currency:     in.Currency,
		AnnualRate:   in.AnnualRate,
		TenureMonths: in.TenureMonths,
		EMIAmount:    domain.CalculateEMI(principal, in.AnnualRate, in.TenureMonths),
		Outstanding:  principal,
		Status:       domain.StatusPending,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, loan); err != nil {
		return domain.Loan{}, err
	}
	return loan, nil
}

// Approve transitions a loan from PENDING to APPROVED (staff action).
func (s *Service) Approve(ctx context.Context, id string, approve bool) (domain.Loan, error) {
	loan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Loan{}, err
	}
	if loan.Status != domain.StatusPending {
		return domain.Loan{}, apierror.ErrConflict("loan is not pending")
	}
	if approve {
		loan.Status = domain.StatusApproved
	} else {
		loan.Status = domain.StatusRejected
	}
	if err := s.repo.Update(ctx, loan); err != nil {
		return domain.Loan{}, err
	}
	return loan, nil
}

// Disburse credits the borrower's account with the principal via the ledger.
func (s *Service) Disburse(ctx context.Context, id string) (domain.Loan, error) {
	loan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Loan{}, err
	}
	if loan.Status != domain.StatusApproved {
		return domain.Loan{}, apierror.ErrConflict("loan must be approved before disbursement")
	}

	funding := fundingAccount(loan.Currency)
	if err := s.ledger.OpenAccount(ctx, funding, loan.Currency); err != nil {
		return domain.Loan{}, ledgerErr()
	}
	entries := []ledgerclient.Entry{
		{AccountID: funding, Direction: "DEBIT", Amount: loan.Principal.String()},
		{AccountID: loan.AccountID, Direction: "CREDIT", Amount: loan.Principal.String()},
	}
	if _, _, err := s.ledger.Post(ctx, uuid.NewString(), "loan-disburse-"+loan.ID, loan.Currency, "loan:"+loan.ID, entries); err != nil {
		return domain.Loan{}, ledgerErr()
	}

	now := time.Now().UTC()
	loan.Status = domain.StatusDisbursed
	loan.DisbursedAt = &now
	if err := s.repo.Update(ctx, loan); err != nil {
		return domain.Loan{}, err
	}
	return loan, nil
}

// Get returns a loan by ID.
func (s *Service) Get(ctx context.Context, id string) (domain.Loan, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByBorrower returns a borrower's loans.
func (s *Service) ListByBorrower(ctx context.Context, borrowerID string) ([]domain.Loan, error) {
	return s.repo.ListByBorrower(ctx, borrowerID)
}

func ledgerErr() error {
	return apierror.New(502, "ledger_unavailable", "ledger service unavailable")
}
