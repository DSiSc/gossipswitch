package transaction

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/statedb-NG/util"
	wallett "github.com/DSiSc/wallet/core/types"
	"math/big"
)

// TxFilter is an implemention of switch message filter,
// switch will use transaction filter to verify transaction message.
type TxFilter struct {
	eventCenter     types.EventCenter
	verifySignature bool
	chainId         uint64
}

// create a new transaction filter instance.
func NewTxFilter(eventCenter types.EventCenter, verifySignature bool, chainId uint64) *TxFilter {
	return &TxFilter{
		eventCenter:     eventCenter,
		verifySignature: verifySignature,
		chainId:         chainId,
	}
}

// Verify verify a switch message whether is validated.
// return nil if message is validated, otherwise return relative error
func (txValidator *TxFilter) Verify(portId int, msg interface{}) error {
	switch msg := msg.(type) {
	case *types.Transaction:
		return txValidator.doVerify(msg)
	default:
		return errors.New("unsupported message type")
	}
}

// do verify operation
func (txValidator *TxFilter) doVerify(tx *types.Transaction) error {
	if txValidator.verifySignature {
		signer := wallett.NewEIP155Signer(big.NewInt(int64(txValidator.chainId)))
		//signer := new(wallett.FrontierSigner)
		from, err := wallett.Sender(signer, tx)
		if nil != err {
			log.Error("Get from by tx's signer failed with %v.", err)
			err := fmt.Errorf("Get from by tx's signer failed with %v ", err)
			txValidator.eventCenter.Notify(types.EventTxVerifyFailed, err)
			return err
		}
		if !bytes.Equal((*tx.Data.From)[:], from.Bytes()) {
			log.Error("Transaction signature verify failed. from=%v, tx.data.from=%v, v=%v", from.String(), util.AddressToHex(*(tx.Data.From)), tx.Data.V)
			err := fmt.Errorf("Transaction signature verify failed ")
			txValidator.eventCenter.Notify(types.EventTxVerifyFailed, err)
			return err
		}
	}
	txValidator.eventCenter.Notify(types.EventTxVerifySucceeded, tx)
	return nil
}
