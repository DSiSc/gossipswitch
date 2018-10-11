package block

import (
	"encoding/json"
	gconf "github.com/DSiSc/craft/config"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/crypto-suite/crypto/sha3"
)

// Sum returns the first 32 bytes of hash of the bz.
func Sum(bz []byte) []byte {
	var alg string
	if value, ok := gconf.GlobalConfig.Load(gconf.HashAlgName); ok {
		alg = value.(string)
	} else {
		alg = "SHA256"
	}
	hasher := sha3.NewHashByAlgName(alg)
	hasher.Write(bz)
	hash := hasher.Sum(nil)
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

type RefAddress struct {
	Addr types.Address
}

func NewRefAddress(addr types.Address) *RefAddress {
	return &RefAddress{Addr: addr}
}

func (self *RefAddress) Address() types.Address {
	return self.Addr
}
