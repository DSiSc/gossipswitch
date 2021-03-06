package block

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	common "github.com/DSiSc/gossipswitch/filter"
	"github.com/DSiSc/repository"
	"sync"
)

// TxFilter is an implemention of switch message filter,
// switch will use transaction filter to verify transaction message.
type BlockFilter struct {
	eventCenter     types.EventCenter
	verifySignature bool
	lock            sync.Mutex
}

// create a new block filter instance.
func NewBlockFilter(eventCenter types.EventCenter, verifySignature bool) *BlockFilter {
	return &BlockFilter{
		eventCenter:     eventCenter,
		verifySignature: verifySignature,
	}
}

// Verify verify a switch message whether is validated.
// return nil if message is validated, otherwise return relative error
func (filter *BlockFilter) Verify(portId int, msg interface{}) error {
	filter.lock.Lock()
	defer filter.lock.Unlock()
	var err error
	switch msg := msg.(type) {
	case *types.Block:
		err = filter.doValidate(msg)
	default:
		log.Error("Invalidate block message ")
		err = errors.New("Invalidate block message ")
	}

	//send verification failed event
	if err != nil {
		log.Debug("Send message verification failed event")
	}
	return err
}

// do verify operation
func (filter *BlockFilter) doValidate(block *types.Block) error {
	log.Debug("Start to validate received block %x", block.HeaderHash)

	// verify block header hash
	blockHash := common.HeaderHash(block)
	if !bytes.Equal(blockHash[:], block.HeaderHash[:]) {
		log.Error("block header's hash %x, is not same with expected %x", blockHash, block.HeaderHash)
		err := fmt.Errorf("block header's hash %x, is not same with expected %x", blockHash, block.HeaderHash)
		filter.eventCenter.Notify(types.EventBlockVerifyFailed, err)
		return err
	}

	// retrieve previous world state
	preBlkHash := block.Header.PrevBlockHash
	bc, err := repository.NewRepositoryByBlockHash(preBlkHash)
	if err != nil {
		log.Error("Failed to validate previous block, as: %v", err)
		err := fmt.Errorf("failed to get previous block state, as:%v", err)
		filter.eventCenter.Notify(types.EventBlockVerifyFailed, err)
		return err
	}

	currentHeight := bc.GetCurrentBlockHeight()
	if currentHeight >= block.Header.Height {
		log.Warn("Local block height %d is bigger than received block %x, height: %d", currentHeight, blockHash, block.Header.Height)
		err := fmt.Errorf("Local block height %d is bigger than received block %x, height: %d ", currentHeight, blockHash, block.Header.Height)
		filter.eventCenter.Notify(types.EventBlockExisted, err)
		return err
	}

	// verify block
	blockValidator := getValidateWorker(bc, block, filter.verifySignature)
	err = blockValidator.VerifyBlock()
	if err != nil {
		log.Error("Validate block failed, as %v", err)
		err := fmt.Errorf("Validate block failed, as %v", err)
		filter.eventCenter.Notify(types.EventBlockVerifyFailed, err)
		return err
	}

	// write block to local database
	return bc.WriteBlockWithReceipts(block, blockValidator.GetReceipts())
}

// get validate worker by previous world state and block
func getValidateWorker(bc *repository.Repository, block *types.Block, verifySignature bool) *Worker {
	return NewWorker(bc, block, verifySignature)
}
