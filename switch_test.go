package gossipswitch

import (
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// mock switch filter
type mockSwitchFiler struct {
}

func (filter *mockSwitchFiler) Verify(msg interface{}) error {
	return nil
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
	var _, err = NewGossipSwitchByType(TxSwitch)
	assert.Nil(err, "FAILED: failed to create GossipSwitch")
	_, err = NewGossipSwitchByType(BlockSwitch)
	assert.Nil(err, "FAILED: failed to create GossipSwitch")
}

// Test get switch in port by id
func Test_InPort(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")
	var localInPort = sw.InPort(LocalInPortId)
	assert.NotNil(localInPort, "FAILED: failed to get local in port")

	var remoteInPort = sw.InPort(LocalInPortId)
	assert.NotNil(remoteInPort, "FAILED: failed to get remote in port.")
}

// Test get switch out port by id
func Test_OutPort(t *testing.T) {
	assert := assert.New(t)
	var sw = NewGossipSwitch(&mockSwitchFiler{})
	assert.NotNil(sw, "FAILED: failed to create GossipSwitch")

	var localOutPort = sw.OutPort(LocalOutPortId)
	assert.NotNil(localOutPort, "FAILED: failed to get local out port")

	var remoteOutPort = sw.OutPort(RemoteOutPortId)
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

	//send message to switch
	txMsg := &types.Transaction{}
	sw.InPort(LocalInPortId).Channel() <- txMsg

	recvMsgChan := make(chan interface{})
	// bind output func to switch out port
	outPort := sw.OutPort(LocalOutPortId)
	outPort.BindToPort(func(msg interface{}) error {
		recvMsgChan <- msg
		return nil
	})

	recvMsg := <-recvMsgChan
	if recvMsg != txMsg {
		t.Error("FAILED: failed to receive the message")
	}
	t.Log("PASS: succed receiving the message")
}

// check switch status
func checkSwitchStatus(t *testing.T, err error, currentStatus uint32, expectStatus uint32) {
	if err != nil || currentStatus != expectStatus {
		t.Error("FAILED: switch current status is not the expected status.")
		panic("switch current status is not the expected status.")
	}
	if currentStatus == 0 {
		t.Log("PASS: succed stopping switch")
	} else {
		t.Log("PASS: succed starting switch")
	}
}
