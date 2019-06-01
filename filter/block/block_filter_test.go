package block

import (
	"errors"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/gossipswitch/filter"
	"github.com/DSiSc/gossipswitch/port"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/repository"
	"github.com/DSiSc/repository/config"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

var mockGenesisBlock = &types.Block{
	Header:     &types.Header{},
	HeaderHash: types.Hash{0xbc, 0xf1, 0xf4, 0x1f, 0xa1, 0x28, 0x66, 0x3d, 0x05, 0x98, 0x1e, 0xf1, 0x55, 0x9e, 0x34, 0x3f, 0x5b, 0xe4, 0x86, 0xd6, 0x58, 0xc8, 0xe3, 0xd8, 0x76, 0x4d, 0xfd, 0xd6, 0x8e, 0xfa, 0xce, 0x12},
}

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
	cfg := config.RepositoryConfig{
		PluginName: repository.PLUGIN_MEMDB,
	}
	eventCenter := mockEventCenter()
	repository.InitRepository(cfg, eventCenter)

	bc, _ := repository.NewLatestStateRepository()
	bc.WriteBlock(mockGenesisBlock)
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
	b.Header.PrevBlockHash = filter.HeaderHash(bc.GetCurrentBlock())
	b.HeaderHash = filter.HeaderHash(b)
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
	monkey.Patch(getValidateWorker, func(bc *repository.Repository, block *types.Block, verifySignature bool) *Worker {
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
	monkey.Patch(getValidateWorker, func(bc *repository.Repository, block *types.Block, verifySignature bool) *Worker {
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
