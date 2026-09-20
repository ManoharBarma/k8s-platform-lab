package money

import "testing"

func TestFromStringAndArithmetic(t *testing.T) {
	a, err := FromString("100.50", "USD")
	if err != nil {
		t.Fatalf("FromString: %v", err)
	}
	b, err := FromString("0.50", "USD")
	if err != nil {
		t.Fatalf("FromString: %v", err)
	}

	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got := sum.String(); got != "101.00 USD" {
		t.Fatalf("Add result = %q, want %q", got, "101.00 USD")
	}

	diff, err := a.Sub(b)
	if err != nil {
		t.Fatalf("Sub: %v", err)
	}
	if got := diff.String(); got != "100.00 USD" {
		t.Fatalf("Sub result = %q, want %q", got, "100.00 USD")
	}
}

func TestCurrencyMismatch(t *testing.T) {
	usd, _ := FromString("10", "USD")
	eur, _ := FromString("10", "EUR")
	if _, err := usd.Add(eur); err == nil {
		t.Fatal("expected currency mismatch error, got nil")
	}
}

func TestGreaterThanOrEqual(t *testing.T) {
	balance, _ := FromString("500.00", "USD")
	withdrawal, _ := FromString("500.00", "USD")
	ok, err := balance.GreaterThanOrEqual(withdrawal)
	if err != nil {
		t.Fatalf("GreaterThanOrEqual: %v", err)
	}
	if !ok {
		t.Fatal("expected balance >= withdrawal")
	}
}
