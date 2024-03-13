package domain

type BasketStatus string

const (
	BasketUnknown     BasketStatus = ""
	BasketIsOpen      BasketStatus = "open"
	BasketIsCancelled BasketStatus = "cancelled"
	BasketIsCheckOut  BasketStatus = "check_out"
)

func (s BasketStatus) String() string {
	switch s {
	case BasketIsOpen, BasketIsCancelled, BasketIsCheckOut:
		return string(s)
	default:
		return ""
	}
}

func ToBasketStatus(status string) BasketStatus {
	switch status {
	case BasketIsOpen.String():
		return BasketIsOpen
	case BasketIsCancelled.String():
		return BasketIsCancelled
	case BasketIsCheckOut.String():
		return BasketIsCheckOut
	default:
		return BasketUnknown
	}
}
