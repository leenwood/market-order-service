package domain_test

import (
	"testing"

	"market-order-service/internal/core/domain"
)

func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		from, to domain.OrderStatus
		role     string
		want     bool
	}{
		{domain.StatusPending, domain.StatusCancelled, "buyer", true},
		{domain.StatusConfirmed, domain.StatusCancelled, "buyer", false},
		{domain.StatusPending, domain.StatusConfirmed, "seller", true},
		{domain.StatusPending, domain.StatusCancelled, "seller", true},
		{domain.StatusConfirmed, domain.StatusShipped, "seller", true},
		{domain.StatusConfirmed, domain.StatusCancelled, "seller", true},
		{domain.StatusShipped, domain.StatusDelivered, "admin", true},
		{domain.StatusShipped, domain.StatusCancelled, "admin", false},
		{domain.StatusDelivered, domain.StatusCancelled, "admin", false},
	}

	for _, tc := range tests {
		got := domain.CanTransitionTo(tc.from, tc.to, tc.role)
		if got != tc.want {
			t.Errorf("CanTransitionTo(%s→%s, %s) = %v, want %v",
				tc.from, tc.to, tc.role, got, tc.want)
		}
	}
}
