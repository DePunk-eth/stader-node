package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/urfave/cli"

	"github.com/stader-labs/stader-node/shared/services"
	"github.com/stader-labs/stader-node/shared/types/api"
	"github.com/stader-labs/stader-node/shared/utils/eth1"
)

func canDelegateUpgrade(c *cli.Context, minipoolAddress common.Address) (*api.CanDelegateUpgradeResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.CanDelegateUpgradeResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get latest delegate address
	latestDelegateAddress, err := rp.GetAddress("rocketMinipoolDelegate", nil)
	if err != nil {
		return nil, err
	}
	response.LatestDelegateAddress = *latestDelegateAddress

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateDelegateUpgradeGas(opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Return response
	return &response, nil

}

func delegateUpgrade(c *cli.Context, minipoolAddress common.Address) (*api.DelegateUpgradeResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.DelegateUpgradeResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

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

	// Upgrade
	hash, err := mp.DelegateUpgrade(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canDelegateRollback(c *cli.Context, minipoolAddress common.Address) (*api.CanDelegateRollbackResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.CanDelegateRollbackResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	rollbackAddress, err := mp.GetPreviousDelegate(nil)
	if err != nil {
		return nil, err
	}
	response.RollbackAddress = rollbackAddress

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateDelegateRollbackGas(opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Return response
	return &response, nil

}

func delegateRollback(c *cli.Context, minipoolAddress common.Address) (*api.DelegateRollbackResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.DelegateRollbackResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

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

	// Rollback
	hash, err := mp.DelegateRollback(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canSetUseLatestDelegate(c *cli.Context, minipoolAddress common.Address, setting bool) (*api.CanSetUseLatestDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.CanSetUseLatestDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateSetUseLatestDelegateGas(setting, opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Return response
	return &response, nil

}

func setUseLatestDelegate(c *cli.Context, minipoolAddress common.Address, setting bool) (*api.SetUseLatestDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.SetUseLatestDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

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

	// Set the new setting
	hash, err := mp.SetUseLatestDelegate(setting, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func getUseLatestDelegate(c *cli.Context, minipoolAddress common.Address) (*api.GetUseLatestDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetUseLatestDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	setting, err := mp.GetUseLatestDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting use latest delegate: %w", err)
	}

	// Return response
	response.Setting = setting
	return &response, nil

}

func getDelegate(c *cli.Context, minipoolAddress common.Address) (*api.GetDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mp.GetDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil

}

func getPreviousDelegate(c *cli.Context, minipoolAddress common.Address) (*api.GetPreviousDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetPreviousDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mp.GetPreviousDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil

}

func getEffectiveDelegate(c *cli.Context, minipoolAddress common.Address) (*api.GetEffectiveDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetEffectiveDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mp.GetEffectiveDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil

}
