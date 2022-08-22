// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var _ = (*callMessageMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (c CallMessage) MarshalJSON() ([]byte, error) {
	type CallMessage struct {
		MsgFrom      common.Address  `json:"from"`
		MsgTo        *common.Address `json:"to"`
		MsgNonce     uint64          `json:"nonce"`
		MsgValue     *hexutil.Big    `json:"gas"`
		MsgGas       uint64          `json:"value"`
		MsgGasPrice  *hexutil.Big    `json:"gas_price"`
		MsgGasFeeCap *hexutil.Big    `json:"gas_fee_cap"`
		MsgGasTipCap *hexutil.Big    `json:"gas_tip_cap"`
		MsgData      []byte          `json:"data"`
	}
	var enc CallMessage
	enc.MsgFrom = c.MsgFrom
	enc.MsgTo = c.MsgTo
	enc.MsgNonce = c.MsgNonce
	enc.MsgValue = (*hexutil.Big)(c.MsgValue)
	enc.MsgGas = c.MsgGas
	enc.MsgGasPrice = (*hexutil.Big)(c.MsgGasPrice)
	enc.MsgGasFeeCap = (*hexutil.Big)(c.MsgGasFeeCap)
	enc.MsgGasTipCap = (*hexutil.Big)(c.MsgGasTipCap)
	enc.MsgData = c.MsgData
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (c *CallMessage) UnmarshalJSON(input []byte) error {
	type CallMessage struct {
		MsgFrom      *common.Address `json:"from"`
		MsgTo        *common.Address `json:"to"`
		MsgNonce     *uint64         `json:"nonce"`
		MsgValue     *hexutil.Big    `json:"gas"`
		MsgGas       *uint64         `json:"value"`
		MsgGasPrice  *hexutil.Big    `json:"gas_price"`
		MsgGasFeeCap *hexutil.Big    `json:"gas_fee_cap"`
		MsgGasTipCap *hexutil.Big    `json:"gas_tip_cap"`
		MsgData      []byte          `json:"data"`
	}
	var dec CallMessage
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.MsgFrom != nil {
		c.MsgFrom = *dec.MsgFrom
	}
	if dec.MsgTo != nil {
		c.MsgTo = dec.MsgTo
	}
	if dec.MsgNonce != nil {
		c.MsgNonce = *dec.MsgNonce
	}
	if dec.MsgValue != nil {
		c.MsgValue = (*big.Int)(dec.MsgValue)
	}
	if dec.MsgGas != nil {
		c.MsgGas = *dec.MsgGas
	}
	if dec.MsgGasPrice != nil {
		c.MsgGasPrice = (*big.Int)(dec.MsgGasPrice)
	}
	if dec.MsgGasFeeCap != nil {
		c.MsgGasFeeCap = (*big.Int)(dec.MsgGasFeeCap)
	}
	if dec.MsgGasTipCap != nil {
		c.MsgGasTipCap = (*big.Int)(dec.MsgGasTipCap)
	}
	if dec.MsgData != nil {
		c.MsgData = dec.MsgData
	}
	return nil
}