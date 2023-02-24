package tokens

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stader-labs/stader-node/stader-lib/stader"
)

// Token balances
type Balances struct {
	ETH  *big.Int `json:"eth"`
	Sd   *big.Int `json:"reth"`
	Ethx *big.Int `json:"rpl"`
}

// Get a token's total supply
func TotalSupply(tokenContract *stader.Erc20TokenContractManager, opts *bind.CallOpts) (*big.Int, error) {
	return tokenContract.Erc20Token.TotalSupply(opts)
}

// Get a token balance
func BalanceOf(tokenContract *stader.Erc20TokenContractManager, address common.Address, opts *bind.CallOpts) (*big.Int, error) {
	return tokenContract.Erc20Token.BalanceOf(opts, address)
}

// Get a spender's Allowance for an address
func Allowance(tokenContract *stader.Erc20TokenContractManager, owner common.Address, spender common.Address, opts *bind.CallOpts) (*big.Int, error) {
	return tokenContract.Erc20Token.Allowance(opts, owner, spender)
}

// Estimate the gas of Transfer
func EstimateTransferGas(tokenContract *stader.Erc20TokenContractManager, to common.Address, amount *big.Int, opts *bind.TransactOpts) (stader.GasInfo, error) {
	return tokenContract.Erc20TokenContract.GetTransactionGasInfo(opts, "Transfer", to, amount)
}

// Transfer tokens to an address
func Transfer(tokenContract *stader.Erc20TokenContractManager, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := tokenContract.Erc20Token.Transfer(opts, to, amount)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.Hash(), nil
}

// Estimate the gas of Approve
func EstimateApproveGas(tokenContract *stader.Erc20TokenContractManager, spender common.Address, amount *big.Int, opts *bind.TransactOpts) (stader.GasInfo, error) {
	return tokenContract.Erc20TokenContract.GetTransactionGasInfo(opts, "Approve", spender, amount)
}

// Approve a token Allowance for a spender
func Approve(tokenContract *stader.Erc20TokenContractManager, spender common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := tokenContract.Erc20Token.Approve(opts, spender, amount)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.Hash(), nil
}

// Estimate the gas of TransferFrom
func EstimateTransferFromGas(tokenContract *stader.Erc20TokenContractManager, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (stader.GasInfo, error) {
	return tokenContract.Erc20TokenContract.GetTransactionGasInfo(opts, "TransferFrom", from, to, amount)
}

// Transfer tokens from a sender to an address
func TransferFrom(tokenContract *stader.Erc20TokenContractManager, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := tokenContract.Erc20Token.TransferFrom(opts, from, to, amount)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.Hash(), nil
}
