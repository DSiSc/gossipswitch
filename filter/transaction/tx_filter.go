package transaction

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	wallett "github.com/DSiSc/wallet/core/types"
)

// TxFilter is an implemention of switch message filter,
// switch will use transaction filter to verify transaction message.
type TxFilter struct {
	eventCenter     types.EventCenter
	verifySignature bool
}

// create a new transaction filter instance.
func NewTxFilter(eventCenter types.EventCenter, verifySignature bool) *TxFilter {
	return &TxFilter{
		eventCenter:     eventCenter,
		verifySignature: verifySignature,
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
		signer := new(wallett.FrontierSigner)
		from, err := wallett.Sender(signer, tx)
		if nil != err {
			log.Error("Get from by tx's signer failed with %v.", err)
			err := fmt.Errorf("Get from by tx's signer failed with %v ", err)
			txValidator.eventCenter.Notify(types.EventTxVerifyFailed, err)
			return err
		}
		if !bytes.Equal((*tx.Data.From)[:], from.Bytes()) {
			log.Error("Transaction signature verify failed.")
			err := fmt.Errorf("Transaction signature verify failed ")
			txValidator.eventCenter.Notify(types.EventTxVerifyFailed, err)
			return err
		}
	}
	txValidator.eventCenter.Notify(types.EventTxVerifySucceeded, tx)
	return nil
}
