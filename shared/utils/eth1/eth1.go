package eth1

// ROCKETPOOL-OWNED

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stader-labs/stader-node/shared/services"
	"github.com/urfave/cli"
	"math/big"
)

// Sets the nonce of the provided transaction options to the latest nonce if requested
func CheckForNonceOverride(c *cli.Context, opts *bind.TransactOpts) error {

	customNonceString := c.GlobalString("nonce")
	if customNonceString != "" {
		customNonce, success := big.NewInt(0).SetString(customNonceString, 0)
		if !success {
			return fmt.Errorf("Invalid nonce: %s", customNonceString)
		}

		// Do a sanity check to make sure the provided nonce is for a pending transaction
		// otherwise the user is burning gas for no reason
		ec, err := services.GetEthClient(c)
		if err != nil {
			return fmt.Errorf("Could not retrieve ETH1 client: %w", err)
		}

		// Make sure it's not higher than the next available nonce
		nextNonceUint, err := ec.PendingNonceAt(context.Background(), opts.From)
		if err != nil {
			return fmt.Errorf("Could not get next available nonce: %w", err)
		}

		nextNonce := big.NewInt(0).SetUint64(nextNonceUint)
		if customNonce.Cmp(nextNonce) == 1 {
			return fmt.Errorf("Can't use nonce %s because it's greater than the next available nonce (%d).", customNonceString, nextNonceUint)
		}

		// Make sure the nonce hasn't already been included in a block
		latestProposedNonceUint, err := ec.NonceAt(context.Background(), opts.From, nil)
		if err != nil {
			return fmt.Errorf("Could not get latest nonce: %w", err)
		}

		latestProposedNonce := big.NewInt(0).SetUint64(latestProposedNonceUint)
		if customNonce.Cmp(latestProposedNonce) == -1 {
			return fmt.Errorf("Can't use nonce %s because it has already been included in a block.", customNonceString)
		}

		// It points to a pending transaction, so this is a valid thing to do
		opts.Nonce = customNonce
	}
	return nil

}
