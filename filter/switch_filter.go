package filter

// Filter is used to verify SwitchMsg
type SwitchFilter interface {
	Verify(msg interface{}) error
}
