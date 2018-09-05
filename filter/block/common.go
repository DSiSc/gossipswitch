package block

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/DSiSc/craft/types"
)

// Sum returns the first 20 bytes of SHA256 of the bz.
func Sum(bz []byte) []byte {
	hash := sha256.Sum256(bz)
	return hash[:types.HashLength]
}

// calculate tx's hash
func TxHash(tx *types.Transaction) (hash types.Hash) {
	jsonByte, _ := json.Marshal(tx)
	sumByte := Sum(jsonByte)
	copy(hash[:], sumByte)
	return
}

// calculate header's hash
func HeaderHash(h *types.Header) (hash types.Hash) {
	jsonByte, _ := json.Marshal(h)
	sumByte := Sum(jsonByte)
	copy(hash[:], sumByte)
	return
}

// calculate block's hash
func BlockHash(block *types.Block) (hash types.Hash) {
	jsonByte, _ := json.Marshal(block)
	sumByte := Sum(jsonByte)
	copy(hash[:], sumByte)
	return
}

type RefAddress struct {
	Addr types.Address
}

func NewRefAddress(addr types.Address) *RefAddress {
	return &RefAddress{Addr: addr}
}

func (self *RefAddress) Address() types.Address {
	return self.Addr
}
