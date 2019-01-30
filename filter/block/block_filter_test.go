package block

import (
	"errors"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/gossipswitch/filter"
	"github.com/DSiSc/gossipswitch/port"
	"github.com/DSiSc/monkey"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

// Test new BlockFiltercd .
func Test_NewBlockFilter(t *testing.T) {
	assert := assert.New(t)
	// init event center
	eventCenter := mockEventCenter()
	var blockFilter = NewBlockFilter(eventCenter, true)
	assert.NotNil(blockFilter, "FAILED: failed to create BlockFilter")
}

// mock block
func mockBlock() *types.Block {
	cfg := config.BlockChainConfig{
		PluginName: blockchain.PLUGIN_MEMDB,
	}
	eventCenter := mockEventCenter()
	blockchain.InitBlockChain(cfg, eventCenter)

	bc, _ := blockchain.NewLatestStateBlockChain()
	b := &types.Block{
		Header: &types.Header{
			ChainID:       1,
			PrevBlockHash: bc.GetCurrentBlock().HeaderHash,
			StateRoot:     bc.IntermediateRoot(false),
			TxRoot:        types.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			ReceiptsRoot:  types.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Height:        1,
			Timestamp:     uint64(time.Date(2018, time.August, 28, 0, 0, 0, 0, time.UTC).Unix()),
			MixDigest:     types.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	b.Header.PrevBlockHash = filter.HeaderHash(bc.GetCurrentBlock().Header)
	b.HeaderHash = filter.HeaderHash(b.Header)
	return b
}

// mock event center
func mockEventCenter() types.EventCenter {
	return &eventCenter{}
}

// Test verify block message.
func TestBlockFilter_Verify(t *testing.T) {
	defer monkey.UnpatchAll()
	assert := assert.New(t)
	// init event center
	eventCenter := mockEventCenter()

	var blockFilter = NewBlockFilter(eventCenter, true)
	assert.NotNil(blockFilter, "FAILED: failed to create BlockFilter")

	tx := &types.Transaction{}
	assert.NotNil(blockFilter.Verify(port.LocalInPortId, tx), "PASS: verify invalid message")
	assert.NotNil(blockFilter.Verify(port.RemoteInPortId, tx), "PASS: verify invalid message")

	block := mockBlock()
	var validateWorker *Worker
	patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(validateWorker), "VerifyBlock", func(self *Worker) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(validateWorker), "GetReceipts", func(self *Worker) types.Receipts {
		return types.Receipts{}
	})
	monkey.Patch(getValidateWorker, func(bc *blockchain.BlockChain, block *types.Block, verifySignature bool) *Worker {
		return validateWorker
	})
	assert.Nil(blockFilter.Verify(port.LocalInPortId, block), "PASS: verify valid block")

	patchGuard.Unpatch()
	monkey.PatchInstanceMethod(reflect.TypeOf(validateWorker), "VerifyBlock", func(self *Worker) error {
		return errors.New("invalid block")
	})
	assert.NotNil(blockFilter.Verify(port.LocalInPortId, block), "PASS: verify invalid block")
	assert.NotNil(blockFilter.Verify(port.RemoteInPortId, block), "PASS: verify invalid block")
}

func TestBlockFilter_Verify2(t *testing.T) {
	defer monkey.UnpatchAll()
	assert := assert.New(t)
	// init event center
	eventCenter := mockEventCenter()

	var blockFilter = NewBlockFilter(eventCenter, true)
	assert.NotNil(blockFilter, "FAILED: failed to create BlockFilter")

	block := mockBlock()
	block.Header.Height = 123
	var validateWorker *Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(validateWorker), "VerifyBlock", func(self *Worker) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(validateWorker), "GetReceipts", func(self *Worker) types.Receipts {
		return types.Receipts{}
	})
	monkey.Patch(getValidateWorker, func(bc *blockchain.BlockChain, block *types.Block, verifySignature bool) *Worker {
		return validateWorker
	})
	assert.NotNil(blockFilter.Verify(port.LocalInPortId, block), "PASS: verify invalid block")
	assert.NotNil(blockFilter.Verify(port.RemoteInPortId, block), "PASS: verify invalid block")
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
