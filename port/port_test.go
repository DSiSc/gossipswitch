package port

import (
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test new a InPort
func Test_NewInPort(t *testing.T) {
	assert := assert.New(t)
	var inPort = NewInPort(LocalInPortId)
	assert.NotNil(inPort, "FAILED: failed to create InPort")

	inPort = NewInPort(RemoteInPortId)
	assert.NotNil(inPort, "FAILED: failed to create InPort")
}

// Test get InPort channel
func TestInPort_Channel(t *testing.T) {
	assert := assert.New(t)
	var inPort = NewInPort(LocalInPortId)
	assert.NotNil(inPort, "FAILED: failed to create InPort")

	inputChannel := inPort.Channel()
	assert.NotNil(inputChannel, "FAILED: failed to get InPort channel")
}

// Test Read message from InPort
func Test_Read(t *testing.T) {
	assert := assert.New(t)
	var inPort = NewInPort(LocalInPortId)
	assert.NotNil(inPort, "FAILED: failed to create InPort")

	txMsg := &types.Transaction{}
	go func() {
		inPort.Channel() <- txMsg
	}()

	recvMsg := <-inPort.Read()
	assert.Equal(recvMsg, txMsg, "FAILED: failed to Read the message")
}

// Test new a OutPort
func Test_NewOutPort(t *testing.T) {
	assert := assert.New(t)
	var outPort = NewOutPort(LocalOutPortId)
	assert.NotNil(outPort, "FAILED: failed to create OutPort")

	outPort = NewOutPort(RemoteOutPortId)
	assert.NotNil(outPort, "FAILED: failed to create OutPort")
}

// Test bind OutPutFunc to OutPort
func TestOutPort_BindToPort(t *testing.T) {
	assert := assert.New(t)
	var outPort = NewOutPort(LocalOutPortId)
	assert.NotNil(outPort, "FAILED: failed to create OutPort")

	outPutFunc := func(msg interface{}) error {
		return nil
	}
	outPort.BindToPort(outPutFunc)

	assert.Condition(
		func() (success bool) {
			return len(outPort.outPutFuncs) == 1
		}, "FAILESD: failed to bind OutPutFunc to OutPort")
}

// Test Write message to OutPort
func TestOutPort_Write(t *testing.T) {
	assert := assert.New(t)
	var outPort = NewOutPort(LocalOutPortId)
	assert.NotNil(outPort, "FAILED: failed to create OutPort")

	var recvMsgChan = make(chan interface{})
	outPutFunc := func(msg interface{}) error {
		recvMsgChan <- msg
		return nil
	}
	outPort.BindToPort(outPutFunc)

	sendMsg := &types.Transaction{}
	outPort.Write(sendMsg)

	recvMsg := <-recvMsgChan
	assert.Equal(recvMsg, sendMsg, "FAILED: failed to Write message to OutPort")
}
