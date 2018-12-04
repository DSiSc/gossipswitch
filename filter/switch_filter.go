package filter

// Filter is used to verify SwitchMsg
type SwitchFilter interface {
	Verify(portId int, msg interface{}) error
}
