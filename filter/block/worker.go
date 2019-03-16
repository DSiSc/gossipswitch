package block

import (
	"bytes"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/evm-NG"
	"github.com/DSiSc/gossipswitch/util"
	vcommon "github.com/DSiSc/validator/common"
	"github.com/DSiSc/validator/tools/merkle_tree"
	"github.com/DSiSc/validator/worker/common"
	wallett "github.com/DSiSc/wallet/core/types"
)

type Worker struct {
	block           *types.Block
	chain           *blockchain.BlockChain
	receipts        types.Receipts
	logs            []*types.Log
	verifySignature bool
}

func NewWorker(chain *blockchain.BlockChain, block *types.Block, verifySignature bool) *Worker {
	return &Worker{
		block:           block,
		chain:           chain,
		verifySignature: verifySignature,
	}
}

func GetTxsRoot(txs []*types.Transaction) types.Hash {
	txHash := make([]types.Hash, 0, len(txs))
	for _, t := range txs {
		txHash = append(txHash, vcommon.TxHash(t))
	}
	txRoot := merkle_tree.ComputeMerkleRoot(txHash)
	return txRoot
}

func (self *Worker) VerifyBlock() error {
	// Get previous block
	previousBlock, err := self.chain.GetBlockByHash(self.block.Header.PrevBlockHash)
	if err != nil {
		log.Error("Failed to validate previous block, as: %v", err)
		return fmt.Errorf("failed to get previous block, as:%v", err)
	}

	// 1. chainID
	if self.block.Header.ChainID != previousBlock.Header.ChainID {
		return fmt.Errorf("wrong Block.Header.ChainID, expected %d, got %d",
			previousBlock.Header.ChainID, self.block.Header.ChainID)
	}
	// 2. hash
	if self.block.Header.PrevBlockHash != previousBlock.HeaderHash {
		return fmt.Errorf("wrong Block.Header.PrevBlockHash, expected %x, got %x",
			previousBlock.HeaderHash, self.block.Header.PrevBlockHash)
	}
	// 3. height
	if self.block.Header.Height != previousBlock.Header.Height+1 {
		return fmt.Errorf("wrong Block.Header.Height, expected %x, got %x",
			previousBlock.Header.Height+1, self.block.Header.Height)
	}
	// 4. txhash
	txsHash := GetTxsRoot(self.block.Transactions)
	if self.block.Header.TxRoot != txsHash {
		return fmt.Errorf("wrong Block.Header.TxRoot, expected %x, got %x",
			txsHash, self.block.Header.TxRoot)
	}
	//5. header hash
	var defaultHash types.Hash
	if !bytes.Equal(defaultHash[:], self.block.HeaderHash[:]) {
		headerHash := vcommon.HeaderHash(self.block)
		if self.block.HeaderHash != headerHash {
			return fmt.Errorf("wrong Block.HeaderHash, expected %x, got %x",
				headerHash, self.block.HeaderHash)
		}
	}

	var (
		receipts types.Receipts
		allLogs  []*types.Log
		gp       = new(common.GasPool).AddGas(uint64(65536))
	)
	// 6. verify every transactions in the block by evm
	for i, tx := range self.block.Transactions {
		self.chain.Prepare(vcommon.TxHash(tx), self.block.Header.PrevBlockHash, i)
		receipt, _, err := self.VerifyTransaction(self.block.Header.CoinBase, gp, self.block.Header, tx, new(uint64))
		if err != nil {
			log.Error("Tx %x verify failed with error %v.", vcommon.TxHash(tx), err)
			return err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	receiptsHash := make([]types.Hash, 0, len(receipts))
	for _, t := range receipts {
		receiptsHash = append(receiptsHash, common.ReceiptHash(t))
		log.Debug("Record tx %x receipt is %x.", t.TxHash, common.ReceiptHash(t))
	}
	receiptHash := merkle_tree.ComputeMerkleRoot(receiptsHash)
	var tempHash types.Hash
	if !bytes.Equal(tempHash[:], self.block.Header.ReceiptsRoot[:]) {
		log.Warn("Receipts root has assigned with %x.", self.block.Header.ReceiptsRoot)
		if !bytes.Equal(receiptHash[:], self.block.Header.ReceiptsRoot[:]) {
			log.Error("Receipts root has assigned with %x, but not consistent with %x.",
				self.block.Header.ReceiptsRoot, receiptHash)
			return fmt.Errorf("receipts hash not consistent")
		}
	} else {
		log.Debug("Assign receipts hash %x to block %d.", receiptHash, self.block.Header.Height)
		self.block.Header.ReceiptsRoot = receiptHash
	}

	// 7. verify state root
	stateRoot := self.chain.IntermediateRoot(false)
	if !bytes.Equal(self.block.Header.StateRoot[:], stateRoot[:]) {
		log.Warn("The calculated StateRoot is inconsistent with the expected value, calculated: %x, expected: %x", stateRoot, self.block.Header.StateRoot)
		return fmt.Errorf("state root is inconsistent")
	}

	// 8. verify digest if it exists
	if !bytes.Equal(defaultHash[:], self.block.Header.MixDigest[:]) {
		digestHash := vcommon.HeaderDigest(self.block.Header)
		if !bytes.Equal(digestHash[:], self.block.Header.MixDigest[:]) {
			log.Error("Block digest not consistent which assignment is [%x], while compute is [%x].",
				self.block.Header.MixDigest, digestHash)
			return fmt.Errorf("digest not consistent")
		}
	}

	self.receipts = receipts
	self.logs = allLogs

	return nil
}

func (self *Worker) VerifyTransaction(author types.Address, gp *common.GasPool, header *types.Header,
	tx *types.Transaction, usedGas *uint64) (*types.Receipt, uint64, error) {
	// verify tx's signature
	if self.verifySignature {
		if self.VerifyTrsSignature(tx) == false {
			log.Error("Transaction signature verify failed.")
			return nil, 0, fmt.Errorf("transaction signature failed")
		}
	}
	context := evm.NewEVMContext(*tx, header, self.chain, author)
	evmEnv := evm.NewEVM(context, self.chain)
	_, gas, failed, err, contractAddr := ApplyTransaction(evmEnv, tx, gp)
	if err != nil {
		log.Error("Apply transaction %x failed with error %v.", vcommon.TxHash(tx), err)
		return nil, 0, err
	}

	root := self.chain.IntermediateRoot(false)
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := common.NewReceipt(vcommon.HashToByte(root), failed, *usedGas)
	receipt.TxHash = vcommon.TxHash(tx)
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if tx.Data.Recipient == nil {
		receipt.ContractAddress = contractAddr
		log.Info("Create contract with address %x within tx %x.", receipt.ContractAddress, receipt.TxHash)
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = self.chain.GetLogs(vcommon.TxHash(tx))
	receipt.Bloom = util.CreateBloom(types.Receipts{receipt})

	return receipt, gas, err
}

func (self *Worker) GetReceipts() types.Receipts {
	log.Debug("Get receipts.")
	return self.receipts
}

func (self *Worker) VerifyTrsSignature(tx *types.Transaction) bool {
	signer := new(wallett.FrontierSigner)
	from, err := wallett.Sender(signer, tx)
	if nil != err {
		log.Error("Get from by tx's %x signer failed with %v.", vcommon.TxHash(tx), err)
		return false
	}
	if !bytes.Equal((*(tx.Data.From))[:], from.Bytes()) {
		log.Error("Transaction signature verify failed, tx.Data.From is %x, while signed from is %x.", *tx.Data.From, from)
		return false
	}
	return true
}
