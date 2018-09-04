package block

import (
	"errors"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/evm-NG"
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
		err = errors.New("unsupported message type")
	}

	//send verification failed event
	if err != nil {
		types.GlobalEventCenter.Notify(types.EventBlockVerifyFailed, err)
	}
	return err
}

// do verify operation
func doValidate(block *types.Block) error {
	//TODO verify over 2/3 signs in block

	// verify block header hash
	hash := HeaderHash(block.Header)
	if hash != block.HeaderHash {
		return errors.New("invalidate block, as actual block header hash is different with expected")
	}

	// retrieve previous block
	preBlkHash := block.Header.PrevBlockHash
	bc, err := blockchain.NewBlockChainByBlockHash(preBlkHash)
	if err != nil {
		return fmt.Errorf("failed to get previous block state, as:%s", err)
	}

	// verify txs in block
	err = executeTxs(bc, block)
	if err != nil {
		return fmt.Errorf("failed to validate block's transactions, as:%s", err)
	}

	// verify state root
	stateRoot := bc.IntermediateRoot(false)
	if stateRoot != block.Header.StateRoot {
		return errors.New("invalidate block, as actual state root is different with expected")
	}

	// write block to local database
	go bc.WriteBlock(block)
	return nil
}

// validate all txs in block
func executeTxs(bc *blockchain.BlockChain, block *types.Block) error {
	gp := new(GasPool).AddGas(uint64(65536))
	for _, tx := range block.Transactions {
		context := evm.NewEVMContext(*tx, block.Header, bc, types.Address{})
		vm := evm.NewEVM(context, bc)
		_, _, _, err := ApplyTransaction(vm, tx, gp)
		if err != nil {
			return err
		}
		return err
	}
	return nil
}
