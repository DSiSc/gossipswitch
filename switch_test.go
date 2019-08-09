package gossipswitch

import (
	"errors"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/gossipswitch/config"
	"github.com/DSiSc/gossipswitch/port"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// mock switch filter
type mockSwitchFiler struct {
}

func (filter *mockSwitchFiler) Verify(portId int, msg interface{}) error {
	return nil
}

// mock switch config
func mockSwitchConfig() *config.SwitchConfig {
	return &config.SwitchConfig{
		VerifySignature: true,
	}
}

// Test new a gossipsiwtch
func Test_NewGossipSwitch(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")
}

// Test new a gossipsiwtch by type
func Test_NewGossipSwitchByType(t *testing.T) {
	assert := assert.New(t)
	var _, err = NewGossipSwitchByType(TxSwitch, &eventCenter{}, mockSwitchConfig())
	assert.Nil(err, "FAILED: failed to create GossipSwitch")
	_, err = NewGossipSwitchByType(BlockSwitch, &eventCenter{}, mockSwitchConfig())
	assert.Nil(err, "FAILED: failed to create GossipSwitch")
}

// Test get switch in port by id
func Test_InPort(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")
	var localInPort = sw.InPort(port.LocalInPortId)
	assert.NotNil(localInPort, "FAILED: failed to get local in port")

	var remoteInPort = sw.InPort(port.LocalInPortId)
	assert.NotNil(remoteInPort, "FAILED: failed to get remote in port.")
}

// Test get switch out port by id
func Test_OutPort(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")

	var localOutPort = sw.OutPort(port.LocalOutPortId)
	assert.NotNil(localOutPort, "FAILED: failed to get local out port")

	var remoteOutPort = sw.OutPort(port.RemoteOutPortId)
	assert.NotNil(remoteOutPort, "FAILED: failed to get remote out port.")
}

// Test start switch
func Test_Start(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")

	err := sw.Start()
	checkSwitchStatus(t, err, sw.isRunning, 1)
}

// Test stop switch
func Test_Stop(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")

	err := sw.Start()
	checkSwitchStatus(t, err, sw.isRunning, 1)

	err = sw.Stop()
	checkSwitchStatus(t, err, sw.isRunning, 0)
}

// Test on receive message
func Test_onRecvMsg(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")

	checkSwitchStatus(t, sw.Start(), sw.isRunning, 1)

	recvMsgChan := make(chan interface{})
	// bind output func to switch out port
	outPort := sw.OutPort(port.LocalOutPortId)
	outPort.BindToPort(func(msg interface{}) error {
		log.Info("received a message")
		recvMsgChan <- msg
		return nil
	})

	//send message to switch
	txMsg := &types.Transaction{}
	sw.InPort(port.LocalInPortId).Channel() <- txMsg

	ticker := time.NewTicker(2 * time.Second)
	select {
	case recvMsg := <-recvMsgChan:
		assert.Equal(txMsg, recvMsg)
	case <-ticker.C:
		assert.Nil(errors.New("failed to receive message"))

	}
}

// check switch status
func checkSwitchStatus(t *testing.T, err error, currentStatus uint32, expectStatus uint32) {
	assert.Equal(t, expectStatus, currentStatus)
	if currentStatus == 0 {
		log.Info("PASS: succed stopping switch")
	} else {
		log.Error("PASS: succed starting switch")
	}
}

type eventCenter struct {
}

// subscriber subscribe specified eventType with eventFunc
func (*eventCenter) Subscribe(eventType types.EventType, eventFunc types.EventFunc) types.Subscriber {
	return nil
}

// subscriber unsubscribe specified eventType
func (*eventCenter) UnSubscribe(eventType types.EventType, subscriber types.Subscriber) (err error) {
	return nil
}

// notify subscriber of eventType
func (*eventCenter) Notify(eventType types.EventType, value interface{}) (err error) {
	return nil
}

// notify specified eventFunc
func (*eventCenter) NotifySubscriber(eventFunc types.EventFunc, value interface{}) {

}

// notify subscriber traversing all events
func (*eventCenter) NotifyAll() (errs []error) {
	return nil
}

// unsubscrible all event
func (*eventCenter) UnSubscribeAll() {
}
