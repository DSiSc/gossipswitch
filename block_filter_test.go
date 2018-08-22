package gossipswitch

import (
	"github.com/DSiSc/producer/common"
	"github.com/DSiSc/txpool/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test new BlockFiltercd .
func Test_NewBlockFilter(t *testing.T) {
	assert := assert.New(t)
	var blockFilter = NewBlockFilter()
	assert.NotNil(blockFilter, "FAILED: failed to create BlockFilter")
}

// Test verify block message.
func Test_BlockFilterVerify(t *testing.T) {
	assert := assert.New(t)
	var blockFilter = NewBlockFilter()
	assert.NotNil(blockFilter, "FAILED: failed to create BlockFilter")

	tx := &types.Transaction{}
	assert.NotNil(blockFilter.Verify(tx), "PASS: verify validated message")

	block := &common.Block{}
	assert.Nil(blockFilter.Verify(block), "PASS: verify in validated message")
}
