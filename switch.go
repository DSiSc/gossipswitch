package gossipswitch

import (
	"errors"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/gossipswitch/config"
	"github.com/DSiSc/gossipswitch/filter"
	"github.com/DSiSc/gossipswitch/filter/block"
	"github.com/DSiSc/gossipswitch/filter/transaction"
	"github.com/DSiSc/gossipswitch/port"
	"sync"
	"sync/atomic"
)

// SwitchType switch type
type SwitchType int

const (
	TxSwitch SwitchType = iota
	BlockSwitch
)

// GossipSwitch is the implementation of gossip switch.
// for gossipswitch, if a validated message is received, it will be broadcasted,
// otherwise it will be dropped.
type GossipSwitch struct {
	switchMtx sync.Mutex
	filter    filter.SwitchFilter
	inPorts   map[int]*port.InPort
	outPorts  map[int]*port.OutPort
	isRunning uint32 // atomic
}

// NewGossipSwitch create a new switch instance with given filter.
// filter is used to verify the received message
func NewGossipSwitch(filter filter.SwitchFilter) *GossipSwitch {
	sw := &GossipSwitch{
		filter:   filter,
		inPorts:  make(map[int]*port.InPort),
		outPorts: make(map[int]*port.OutPort),
	}
	sw.initPort()
	return sw
}

// NewGossipSwitchByType create a new switch instance by type.
// switchType is used to specify the switch type
func NewGossipSwitchByType(switchType SwitchType, eventCenter types.EventCenter, switchConfig *config.SwitchConfig) (*GossipSwitch, error) {
	var msgFilter filter.SwitchFilter
	switch switchType {
	case TxSwitch:
		log.Info("New transaction switch")
		msgFilter = transaction.NewTxFilter(eventCenter, switchConfig.VerifySignature)
	case BlockSwitch:
		log.Info("New block switch")
		msgFilter = block.NewBlockFilter(eventCenter, switchConfig.VerifySignature)
	default:
		log.Error("Unsupported switch type")
		return nil, errors.New("Unsupported switch type ")
	}
	sw := &GossipSwitch{
		filter:   msgFilter,
		inPorts:  make(map[int]*port.InPort),
		outPorts: make(map[int]*port.OutPort),
	}
	sw.initPort()
	return sw, nil
}

// init switch's port.InPort and port.OutPort
func (sw *GossipSwitch) initPort() {
	log.Info("Init switch's ports")
	sw.inPorts[port.LocalInPortId] = port.NewInPort(port.LocalInPortId)
	sw.inPorts[port.RemoteInPortId] = port.NewInPort(port.RemoteInPortId)
	sw.outPorts[port.LocalOutPortId] = port.NewOutPort(port.LocalOutPortId)
	sw.outPorts[port.RemoteOutPortId] = port.NewOutPort(port.RemoteOutPortId)
}

// port.InPort get switch's in port by port id, return nil if there is no port with specific id.
func (sw *GossipSwitch) InPort(portId int) *port.InPort {
	log.Debug("Get switch %v in port", portId)
	return sw.inPorts[portId]
}

// port.InPort get switch's out port by port id, return nil if there is no port with specific id.
func (sw *GossipSwitch) OutPort(portId int) *port.OutPort {
	log.Debug("Get switch %v out port", portId)
	return sw.outPorts[portId]
}

// Start start the switch. Once started, switch will receive message from in port, and broadcast to
// out port
func (sw *GossipSwitch) Start() error {
	log.Info("Begin starting switch")
	if atomic.CompareAndSwapUint32(&sw.isRunning, 0, 1) {
		for _, inPort := range sw.inPorts {
			go sw.receiveRoutine(inPort)
		}
		log.Info("Start switch success")
		return nil
	}
	log.Error("Switch already started")
	return errors.New("switch already started")
}

// Stop stop the switch. Once stopped, switch will stop to receive and broadcast message
func (sw *GossipSwitch) Stop() error {
	log.Info("Begin stopping switch")
	if atomic.CompareAndSwapUint32(&sw.isRunning, 1, 0) {
		log.Info("Stop switch success")
		return nil
	}
	log.Error("Switch already stopped")
	return errors.New("switch already stopped")
}

// IsRunning is used to query switch's current status. Return true if running, otherwise false
func (sw *GossipSwitch) IsRunning() bool {
	return atomic.LoadUint32(&sw.isRunning) == 1
}

// listen to receive message from the in port
func (sw *GossipSwitch) receiveRoutine(inPort *port.InPort) {
	for {
		select {
		case msg := <-inPort.Read():
			sw.onRecvMsg(inPort.PortId(), msg)
		}

		//check switch status
		if !sw.IsRunning() {
			break
		}
	}
}

// deal with the received message.
func (sw *GossipSwitch) onRecvMsg(portId int, msg interface{}) {
	//TODO log.Debug("Received a message %v from port.InPort", msg)
	if err := sw.filter.Verify(portId, msg); err == nil {
		sw.broadCastMsg(msg)
	}
}

// broadcast the validated message to all out ports.
func (sw *GossipSwitch) broadCastMsg(msg interface{}) error {
	log.Debug("Broadcast message %v to port.OutPorts", msg)
	for _, outPort := range sw.outPorts {
		go outPort.Write(msg)
	}
	return nil
}
