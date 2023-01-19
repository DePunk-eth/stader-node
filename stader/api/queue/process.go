package queue

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/stader-labs/stader-node/shared/services"
	"github.com/stader-labs/stader-node/shared/types/api"
	"github.com/stader-labs/stader-node/shared/utils/eth1"
)

func canProcessQueue(c *cli.Context) (*api.CanProcessQueueResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanProcessQueueResponse{}

	// Data
	var wg errgroup.Group
	var nextMinipoolCapacity *big.Int
	var depositPoolBalance *big.Int

	// Check deposit assignments are enabled
	wg.Go(func() error {
		assignDepositsEnabled, err := protocol.GetAssignDepositsEnabled(rp, nil)
		if err == nil {
			response.AssignDepositsDisabled = !assignDepositsEnabled
		}
		return err
	})

	// Get next available minipool capacity
	wg.Go(func() error {
		var err error
		nextMinipoolCapacity, err = minipool.GetQueueNextCapacity(rp, nil)
		return err
	})

	// Get deposit pool balance
	wg.Go(func() error {
		var err error
		depositPoolBalance, err = deposit.GetBalance(rp, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := deposit.EstimateAssignDepositsGas(rp, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check next minipool capacity & deposit pool balance
	response.NoMinipoolsAvailable = (nextMinipoolCapacity.Cmp(big.NewInt(0)) == 0)
	response.InsufficientDepositBalance = (depositPoolBalance.Cmp(nextMinipoolCapacity) < 0)

	// Update & return response
	response.CanProcess = !(response.AssignDepositsDisabled || response.NoMinipoolsAvailable || response.InsufficientDepositBalance)
	return &response, nil

}

func processQueue(c *cli.Context) (*api.ProcessQueueResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProcessQueueResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Process queue
	hash, err := deposit.AssignDeposits(rp, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
