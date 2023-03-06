package node

import (
	"fmt"
	"github.com/stader-labs/stader-node/stader-lib/stader"
	"math/big"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/legacy/v1.0.0/rewards"
	"github.com/stader-labs/stader-node/shared/services"
	"github.com/stader-labs/stader-node/shared/types/api"
	"github.com/stader-labs/stader-node/shared/utils/eth1"
)

func canNodeClaimRpl(c *cli.Context) (*api.CanNodeClaimRplResponse, error) {

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
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeClaimRplResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check for rewards
	legacyClaimNodeAddress := cfg.Smartnode.GetLegacyClaimNodeAddress()
	legacyRewardsPoolAddress := cfg.Smartnode.GetLegacyRewardsPoolAddress()
	rewardsAmountWei, err := rewards.GetNodeClaimRewardsAmount(rp, nodeAccount.Address, nil, &legacyClaimNodeAddress)
	if err != nil {
		return nil, fmt.Errorf("Error getting RPL rewards amount: %w", err)
	}
	response.RplAmount = rewardsAmountWei

	// Don't claim unless the oDAO has claimed first (prevent known issue yet to be patched in smart contracts)
	trustedNodeClaimed, err := rewards.GetTrustedNodeTotalClaimed(rp, nil, &legacyRewardsPoolAddress)
	if err != nil {
		return nil, fmt.Errorf("Error checking if trusted node has already minted RPL: %w", err)
	}
	if trustedNodeClaimed.Cmp(big.NewInt(0)) == 0 {
		response.RplAmount = big.NewInt(0)
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := rewards.EstimateClaimNodeRewardsGas(rp, opts, &legacyClaimNodeAddress)
	if err != nil {
		return nil, fmt.Errorf("Could not estimate the gas required to claim RPL: %w", err)
	}
	response.GasInfo = stader.GasInfo(gasInfo)

	return &response, nil
}

func nodeClaimRpl(c *cli.Context) (*api.NodeClaimRplResponse, error) {

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
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeClaimRplResponse{}

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

	// Claim rewards
	legacyClaimNodeAddress := cfg.Smartnode.GetLegacyClaimNodeAddress()
	hash, err := rewards.ClaimNodeRewards(rp, opts, &legacyClaimNodeAddress)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
