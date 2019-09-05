package port

import (
	"github.com/DSiSc/craft/log"
	"sync"
)

// common const value
const (
	LocalInPortId   = 0 //Local InPort ID, receive the message from local
	RemoteInPortId  = 1 //Remote InPort ID, receive the message from remote
	LocalOutPortId  = 0 //Local OutPort ID
	RemoteOutPortId = 1 //Remote OutPort ID
)

// state is used to record switch port state. e.g., message statistics
type state struct {
	InCount  uint
	OutCount uint
}

// InPort is switch in port. Message will be send to InPort, and then switch Read the message from InPort
type InPort struct {
	id      int
	state   state
	channel chan interface{}
}

// create a new in port instance
func NewInPort(id int) *InPort {
	return &InPort{
		id:      id,
		state:   state{},
		channel: make(chan interface{}),
	}
}

// Channel return the input channel of this InPort
func (inPort *InPort) Channel() chan<- interface{} {
	return inPort.channel
}

// PortId return this port's id
func (inPort *InPort) PortId() int {
	return inPort.id
}

// Read message from this InPort, will be blocked until the message arrives.
func (inPort *InPort) Read() <-chan interface{} {
	return inPort.channel
}

// OutPutFunc is binded to switch out port, and OutPort will call OutPutFunc when receive a message from switch
type OutPutFunc func(msg interface{}) error

// InPort is switch out port. Switch will broadcast message to out port
type OutPort struct {
	id          int
	outPortMtx  sync.Mutex
	state       state
	outPutFuncs []OutPutFunc
}

// create a new out port instance
func NewOutPort(id int) *OutPort {
	return &OutPort{
		id:    id,
		state: state{},
	}
}

// BindToPort bind a new OutPutFunc to this OutPort. Return error if bind failed
func (outPort *OutPort) BindToPort(outPutFunc OutPutFunc) error {
	log.Info("Bind OutPutFunc to OutPort")
	outPort.outPortMtx.Lock()
	defer outPort.outPortMtx.Unlock()
	outPort.outPutFuncs = append(outPort.outPutFuncs, outPutFunc)
	return nil
}

// PortId return this port's id
func (outPort *OutPort) PortId() int {
	return outPort.id
}

// Write message to this OutPort
func (outPort *OutPort) Write(msg interface{}) error {
	outPort.outPortMtx.Lock()
	defer outPort.outPortMtx.Unlock()
	for _, outPutFunc := range outPort.outPutFuncs {
		outPutFunc(msg)
	}
	return nil
}
