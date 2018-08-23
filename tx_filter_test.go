package gossipswitch

import (
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test new TxFilter
func Test_NewTxFilter(t *testing.T) {
	assert := assert.New(t)
	var txFilter = NewTxFilter()
	assert.NotNil(txFilter, "FAILED: failed to create TxFilter")
}

// Test verify transaction message.
func Test_TxFilterVerify(t *testing.T) {
	assert := assert.New(t)
	var txFilter = NewTxFilter()
	assert.NotNil(txFilter, "FAILED: failed to create TxFilter")

	tx := &types.Transaction{}
	assert.Nil(txFilter.Verify(tx), "PASS: verify validated message")

	block := &types.Block{}
	assert.NotNil(txFilter.Verify(block), "PASS: verify in validated message")
}
