package filter

import (
	"github.com/DSiSc/gossipswitch/common"
)

// Filter is used to verify SwitchMsg
type SwitchFilter interface {
	Verify(msg common.SwitchMsg) error
}
