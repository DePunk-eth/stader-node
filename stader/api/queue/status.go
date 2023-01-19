package queue

import (
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/stader-labs/stader-node/shared/services"
	"github.com/stader-labs/stader-node/shared/types/api"
)

func getStatus(c *cli.Context) (*api.QueueStatusResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.QueueStatusResponse{}

	// Sync
	var wg errgroup.Group

	// Get deposit pool balance
	wg.Go(func() error {
		var err error
		response.DepositPoolBalance, err = deposit.GetBalance(rp, nil)
		return err
	})

	// Get minipool queue length
	wg.Go(func() error {
		var err error
		response.MinipoolQueueLength, err = minipool.GetQueueTotalLength(rp, nil)
		return err
	})

	// Get minipool queue capacity
	wg.Go(func() error {
		var err error
		response.MinipoolQueueCapacity, err = minipool.GetQueueTotalCapacity(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
