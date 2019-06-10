package scheduling

type MaxBookingsError struct{}

func (e MaxBookingsError) Error() string {
	return "Max number of bookings reached per week"
}
