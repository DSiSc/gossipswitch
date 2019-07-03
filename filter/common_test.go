package filter

import (
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

var MockHash = types.Hash{
	0x1d, 0xcf, 0x7, 0xba, 0xfc, 0x42, 0xb0, 0x8d, 0xfd, 0x23, 0x9c, 0x45, 0xa4, 0xb9, 0x38, 0xd,
	0x8d, 0xfe, 0x5d, 0x6f, 0xa7, 0xdb, 0xd5, 0x50, 0xc9, 0x25, 0xb1, 0xb3, 0x4, 0xdc, 0xc5, 0x1c,
}

func MockBlock() *types.Block {
	return &types.Block{
		Header: &types.Header{
			ChainID:       1,
			PrevBlockHash: MockHash,
			StateRoot:     MockHash,
			TxRoot:        MockHash,
			ReceiptsRoot:  MockHash,
			Height:        1,
			Timestamp:     uint64(time.Date(2018, time.August, 28, 0, 0, 0, 0, time.UTC).Unix()),
		},
		Transactions: make([]*types.Transaction, 0),
	}
}

var MockBlockHash = types.Hash{
	0x44, 0x49, 0xc9, 0xd9, 0xa3, 0x6a, 0x96, 0xeb, 0x28, 0xc9, 0xe1, 0x80, 0x99, 0x0, 0x5c, 0xcc, 0x65, 0x94, 0x2d, 0x5f, 0x88, 0xdd, 0x1a, 0x5a, 0x9c, 0xcf, 0xff, 0x1, 0xaa, 0x2, 0xf1, 0x76,
}

func TestHeaderHash(t *testing.T) {
	block := MockBlock()
	hash := HeaderHash(block)
	assert.Equal(t, MockBlockHash, hash)
}

// New a transaction
func mockTransaction(nonce uint64, to *types.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, from *types.Address) *types.Transaction {
	d := types.TxData{
		AccountNonce: nonce,
		Recipient:    to,
		From:         from,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &types.Transaction{Data: d}
}

func TestTxHash(t *testing.T) {
	assert := assert.New(t)
	b := types.Address{
		0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
		0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
	}
	emptyTx := mockTransaction(
		0,
		&b,
		big.NewInt(0),
		0,
		big.NewInt(0),
		b[:10],
		&b,
	)
	exceptHash := types.Hash{
		0x63, 0xa2, 0xa4, 0x4, 0x8d, 0x2c, 0xe4, 0xe8, 0x95, 0xd9, 0x24, 0x21, 0xb3, 0xc7, 0x36, 0xa8, 0xed, 0xf0, 0x83, 0xb7, 0xab, 0x9d, 0xf6, 0xee, 0x7f, 0x4b, 0x57, 0x19, 0xf9, 0x78, 0xef, 0x93,
	}
	txHash := TxHash(emptyTx)
	assert.Equal(exceptHash, txHash)

	exceptHash1 := TxHash(emptyTx)
	assert.Equal(exceptHash, exceptHash1)
}
