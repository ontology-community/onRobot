/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	common2 "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"math/rand"
)

var (
	ONT_CONTRACT_VERSION = byte(0)
	ONG_CONTRACT_VERSION = byte(0)

	ONT_CONTRACT_ADDRESS, _ = common.AddressFromHexString("0100000000000000000000000000000000000000")
	ONG_CONTRACT_ADDRESS, _ = common.AddressFromHexString("0200000000000000000000000000000000000000")
)

// Uint256FromHexString
func Uint256FromHexString(hex string) (common.Uint256, error) {
	return common.Uint256FromHexString(hex)
}

// AddressFromBase58
func AddressFromBase58(str string) (common.Address, error) {
	return common.AddressFromBase58(str)
}

// GetBlockHeight
func GetBlockCurrentHeight(addr string) (uint64, error) {
	data, ontErr := sendRpcRequest(addr, "getblockcount", []interface{}{})
	if ontErr != nil {
		return 0, ontErr.Error
	}
	num := uint64(0)
	if err := json.Unmarshal(data, &num); err != nil {
		return 0, fmt.Errorf("json.Unmarshal:%s error:%s", data, err)
	}
	return num, nil
}

// GetBalance
func GetBalance(rpc string, addr common.Address) (*common2.BalanceOfRsp, error) {
	data, ontErr := sendRpcRequest(rpc, "getbalance", []interface{}{addr.ToBase58()})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	resp := &common2.BalanceOfRsp{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("json.Unmarshal:%s error:%s", data, err)
	}
	return resp, nil
}

// GetTxByHash
func GetTxByHash(addr string, hash common.Uint256) (*types.Transaction, error) {
	data, ontErr := sendRpcRequest(addr, "getrawtransaction", []interface{}{hash.ToHexString()})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	return GetTransactionFromBytes(data)
}

func GetTransactionFromBytes(data []byte) (*types.Transaction, error) {
	hexStr := ""
	err := json.Unmarshal(data, &hexStr)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	txData, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return types.TransactionFromRawBytes(txData)
}

// RecoverAccount
func RecoverAccount(walletPath, walletPwd string) (*account.Account, error) {
	wallet, err := account.Open(walletPath)
	if err != nil {
		return nil, err
	}
	return wallet.GetDefaultAccount([]byte(walletPwd))
}

// Sign
func Sign(kp keypair.PrivateKey, hash []byte) (*signature.Signature, error) {
	return signature.Sign(signature.SHA256withECDSA, kp, hash, nil)
}

// TransferOntTx generate transfer ont transaction
func TransferOntTx(gasPrice, gasLimit uint64,
	payer *account.Account,
	to common.Address,
	amount uint64) (*types.Transaction, error) {

	mutableTransaction, err := newTransferTransaction(gasPrice, gasLimit, payer.Address, to, amount)
	if err != nil {
		return nil, err
	}
	mutableTransaction.Payer = payer.Address
	err = signToTransaction(mutableTransaction, payer)
	if err != nil {
		return nil, err
	}
	return mutableTransaction.IntoImmutable()
}

func signToTransaction(tx *types.MutableTransaction, signer *account.Account) error {
	txHash := tx.Hash()
	sig, err := Sign(signer.PrivateKey, txHash.ToArray())
	if err != nil {
		return fmt.Errorf("sign error:%s", err)
	}
	sigData, err := signature.Serialize(sig)
	if err != nil {
		return fmt.Errorf("signature.Serialize error:%s", err)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		PubKeys: []keypair.PublicKey{signer.PubKey()},
		M:       1,
		SigData: [][]byte{sigData},
	})
	return nil
}

func newTransferTransaction(gasPrice, gasLimit uint64, from, to common.Address, amount uint64) (*types.MutableTransaction, error) {
	states := []*ont.State{&ont.State{
		From:  from,
		To:    to,
		Value: amount,
	}}
	params := []interface{}{states}
	invokeCode, err := utils.BuildNativeInvokeCode(ONT_CONTRACT_ADDRESS, ONT_CONTRACT_VERSION, ont.TRANSFER_NAME, params)
	if err != nil {
		return nil, fmt.Errorf("BuildNativeInvokeCode error:%s", err)
	}
	invokePayload := &payload.InvokeCode{Code: invokeCode}
	mutableTx := &types.MutableTransaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.InvokeNeo,
		Nonce:    rand.Uint32(),
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	return mutableTx, nil
}
