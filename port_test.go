package gossipswitch

import (
	"github.com/DSiSc/txpool/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test new a InPort
func Test_NewInPort(t *testing.T) {
	assert := assert.New(t)
	var inPort = newInPort()
	assert.NotNil(inPort, "FAILED: failed to create InPort")
}

// Test get InPort channel
func TestInPort_Channel(t *testing.T) {
	assert := assert.New(t)
	var inPort = newInPort()
	assert.NotNil(inPort, "FAILED: failed to create InPort")

	inputChannel := inPort.Channel()
	assert.NotNil(inputChannel, "FAILED: failed to get InPort channel")
}

// Test read message from InPort
func Test_Read(t *testing.T) {
	assert := assert.New(t)
	var inPort = newInPort()
	assert.NotNil(inPort, "FAILED: failed to create InPort")

	txMsg := &types.Transaction{}
	go func() {
		inPort.Channel() <- txMsg
	}()

	recvMsg := <-inPort.read()
	assert.Equal(recvMsg, txMsg, "FAILED: failed to read the message")
}

// Test new a OutPort
func Test_NewOutPort(t *testing.T) {
	assert := assert.New(t)
	var outPort = newOutPort()
	assert.NotNil(outPort, "FAILED: failed to create OutPort")
}

// Test bind OutPutFunc to OutPort
func TestOutPort_BindToPort(t *testing.T) {
	assert := assert.New(t)
	var outPort = newOutPort()
	assert.NotNil(outPort, "FAILED: failed to create OutPort")

	outPutFunc := func(msg SwitchMsg) error {
		return nil
	}
	outPort.BindToPort(outPutFunc)

	assert.Condition(
		func() (success bool) {
			return len(outPort.outPutFuncs) == 1
		}, "FAILESD: failed to bind OutPutFunc to OutPort")
}

// Test write message to OutPort
func TestOutPort_Write(t *testing.T) {
	assert := assert.New(t)
	var outPort = newOutPort()
	assert.NotNil(outPort, "FAILED: failed to create OutPort")

	var recvMsgChan = make(chan SwitchMsg)
	outPutFunc := func(msg SwitchMsg) error {
		recvMsgChan <- msg
		return nil
	}
	outPort.BindToPort(outPutFunc)

	sendMsg := &types.Transaction{}
	outPort.write(sendMsg)

	recvMsg := <-recvMsgChan
	assert.Equal(recvMsg, sendMsg, "FAILED: failed to write message to OutPort")
}
