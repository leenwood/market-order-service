package domain

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	}
	return false
}

// CanTransitionTo returns true if transitioning from→to is allowed for the given role.
// role: "buyer", "seller", "admin"
func CanTransitionTo(from, to OrderStatus, role string) bool {
	switch role {
	case "buyer":
		return from == StatusPending && to == StatusCancelled
	case "seller", "admin":
		switch from {
		case StatusPending:
			return to == StatusConfirmed || to == StatusCancelled
		case StatusConfirmed:
			return to == StatusShipped || to == StatusCancelled
		case StatusShipped:
			return to == StatusDelivered
		}
	}
	return false
}
