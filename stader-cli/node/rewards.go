package node

import (
	"fmt"
	"time"

	"github.com/urfave/cli"

	"github.com/stader-labs/stader-node/shared/services/stader"
	cliutils "github.com/stader-labs/stader-node/shared/utils/cli"
)

func getRewards(c *cli.Context) error {

	// Get RP client
	staderClient, err := stader.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer staderClient.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(staderClient)
	if err != nil {
		return err
	}

	// Get node RPL rewards status
	rewards, err := staderClient.NodeRewards()
	if err != nil {
		return err
	}

	if !rewards.Registered {
		fmt.Printf("This node is not currently registered.\n")
		return nil
	}

	fmt.Printf("%sNOTE: Legacy rewards from pre-Redstone are temporarily not being included in the below figures. They will be added back in a future release. We apologize for the inconvenience!%s\n\n", colorYellow, colorReset)

	fmt.Println("=== ETH ===")
	fmt.Printf("You have earned %.4f ETH from the Beacon Chain (including your commissions) so far.\n", rewards.BeaconRewards)
	fmt.Printf("You have claimed %.4f ETH from the Smoothing Pool.\n", rewards.CumulativeEthRewards)
	fmt.Printf("You still have %.4f ETH in unclaimed Smoothing Pool rewards.\n", rewards.UnclaimedEthRewards)

	nextRewardsTime := rewards.LastCheckpoint.Add(rewards.RewardsInterval)
	nextRewardsTimeString := cliutils.GetDateTimeString(uint64(nextRewardsTime.Unix()))
	timeToCheckpointString := time.Until(nextRewardsTime).Round(time.Second).String()

	// Assume 365 days in a year, 24 hours per day
	rplApr := rewards.EstimatedRewards / rewards.TotalRplStake / rewards.RewardsInterval.Hours() * (24 * 365) * 100

	fmt.Println("\n=== RPL ===")
	fmt.Printf("The current rewards cycle started on %s.\n", cliutils.GetDateTimeString(uint64(rewards.LastCheckpoint.Unix())))
	fmt.Printf("It will end on %s (%s from now).\n", nextRewardsTimeString, timeToCheckpointString)

	if rewards.UnclaimedRplRewards > 0 {
		fmt.Printf("You currently have %f unclaimed RPL from staking rewards.\n", rewards.UnclaimedRplRewards)
	}
	if rewards.UnclaimedTrustedRplRewards > 0 {
		fmt.Printf("You currently have %f unclaimed RPL from Oracle DAO duties.\n", rewards.UnclaimedTrustedRplRewards)
	}

	fmt.Println()
	fmt.Printf("Your estimated RPL staking rewards for this cycle: %f RPL (this may change based on network activity).\n", rewards.EstimatedRewards)
	fmt.Printf("Based on your current total stake of %f RPL, this is approximately %.2f%% APR.\n", rewards.TotalRplStake, rplApr)
	fmt.Printf("Your node has received %f RPL staking rewards in total.\n", rewards.CumulativeRplRewards)

	if rewards.Trusted {
		rplTrustedApr := rewards.EstimatedTrustedRplRewards / rewards.TrustedRplBond / rewards.RewardsInterval.Hours() * (24 * 365) * 100

		fmt.Println()
		fmt.Printf("You will receive an estimated %f RPL in rewards for Oracle DAO duties (this may change based on network activity).\n", rewards.EstimatedTrustedRplRewards)
		fmt.Printf("Based on your bond of %f RPL, this is approximately %.2f%% APR.\n", rewards.TrustedRplBond, rplTrustedApr)
		fmt.Printf("Your node has received %f RPL Oracle DAO rewards in total.\n", rewards.CumulativeTrustedRplRewards)
	}

	fmt.Println()
	fmt.Println("You may claim these rewards at any time. You no longer need to claim them within this interval.")

	// Return
	return nil

}
