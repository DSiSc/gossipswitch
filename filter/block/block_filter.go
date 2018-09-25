package block

import (
	"errors"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/worker"
)

// TxFilter is an implemention of switch message filter,
// switch will use transaction filter to verify transaction message.
type BlockFilter struct {
}

// create a new block filter instance.
func NewBlockFilter() *BlockFilter {
	return &BlockFilter{}
}

// Verify verify a switch message whether is validated.
// return nil if message is validated, otherwise return relative error
func (blockValidator *BlockFilter) Verify(msg interface{}) error {
	var err error
	switch msg := msg.(type) {
	case *types.Block:
		err = doValidate(msg)
	default:
		log.Error("Invalidate block message ")
		err = errors.New("Invalidate block message ")
	}

	//send verification failed event
	if err != nil {
		log.Debug("Send message verification failed event")
		types.GlobalEventCenter.Notify(types.EventBlockVerifyFailed, err)
	}
	return err
}

// do verify operation
func doValidate(block *types.Block) error {
	log.Debug("Start to validate received block %v", block)
	// retrieve previous world state
	preBlkHash := block.Header.PrevBlockHash
	bc, err := blockchain.NewBlockChainByBlockHash(preBlkHash)
	if err != nil {
		log.Error("Failed to validate previous block, as: %v", err)
		return fmt.Errorf("failed to get previous block state, as:%v", err)
	}

	// verify block
	blockValidator := getValidateWorker(bc, block)
	err = blockValidator.VerifyBlock()
	if err != nil {
		log.Error("Validate block failed, as %v", err)
		return err
	}

	// write block to local database
	go bc.WriteBlock(block)
	return nil
}

// get validate worker by previous world state and block
func getValidateWorker(bc *blockchain.BlockChain, block *types.Block) *worker.Worker {
	return worker.NewWorker(bc, block)
}
