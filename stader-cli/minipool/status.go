package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/stader-labs/stader-node/shared/services/stader"
	"github.com/stader-labs/stader-node/shared/types/api"
	cliutils "github.com/stader-labs/stader-node/shared/utils/cli"
	"github.com/stader-labs/stader-node/shared/utils/hex"
	"github.com/stader-labs/stader-node/shared/utils/math"
)

const colorReset string = "\033[0m"
const colorRed string = "\033[31m"
const colorYellow string = "\033[33m"

func getStatus(c *cli.Context) error {

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

	// Get minipool statuses
	status, err := staderClient.MinipoolStatus()
	if err != nil {
		return err
	}

	// Get minipools by status
	statusMinipools := map[string][]api.MinipoolDetails{}
	refundableMinipools := []api.MinipoolDetails{}
	withdrawableMinipools := []api.MinipoolDetails{}
	closeableMinipools := []api.MinipoolDetails{}
	finalisedMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {

		if !minipool.Finalised {
			// Add to status list
			statusName := minipool.Status.Status.String()
			if _, ok := statusMinipools[statusName]; !ok {
				statusMinipools[statusName] = []api.MinipoolDetails{}
			}
			statusMinipools[statusName] = append(statusMinipools[statusName], minipool)

			// Add to actionable lists
			if minipool.RefundAvailable {
				refundableMinipools = append(refundableMinipools, minipool)
			}
			if minipool.WithdrawalAvailable {
				withdrawableMinipools = append(withdrawableMinipools, minipool)
			}
			if minipool.CloseAvailable {
				closeableMinipools = append(closeableMinipools, minipool)
			}
		} else {
			finalisedMinipools = append(finalisedMinipools, minipool)
		}

	}

	// Print minipool details by status
	if len(status.Minipools) == 0 {
		fmt.Println("The node does not have any minipools yet.")
	}
	for _, statusName := range types.MinipoolStatuses {
		minipools, ok := statusMinipools[statusName]
		if !ok {
			continue
		}

		// Minipool status count & description
		fmt.Printf("%d %s minipool(s):\n", len(minipools), statusName)
		if statusName == "Withdrawable" {
			fmt.Println("(Withdrawal may not be available until after withdrawal delay)")
		}
		fmt.Println("")

		// Minipools
		for _, minipool := range minipools {
			printMinipoolDetails(minipool, status.LatestDelegate)
		}

		fmt.Println("")
	}

	// Handle finalized minipools
	fmt.Printf("%d finalized minipool(s):\n", len(finalisedMinipools))
	fmt.Println("")

	// Minipools
	for _, minipool := range finalisedMinipools {
		printMinipoolDetails(minipool, status.LatestDelegate)
	}

	fmt.Println("")

	// Print actionable minipool details
	if len(refundableMinipools) > 0 {
		fmt.Printf("%d minipool(s) have refunds available:\n", len(refundableMinipools))
		for _, minipool := range refundableMinipools {
			fmt.Printf("- %s (%.6f ETH to claim)\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.RefundBalance), 6))
		}
		fmt.Println("")
	}
	if len(closeableMinipools) > 0 {
		fmt.Printf("%d dissolved minipool(s) can be closed once Beacon Chain withdrawals are enabled:\n", len(closeableMinipools))
		for _, minipool := range closeableMinipools {
			fmt.Printf("- %s (%.6f ETH to claim)\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.DepositBalance), 6))
		}
		fmt.Println("")
	}

	// Return
	return nil

}

func printMinipoolDetails(minipool api.MinipoolDetails, latestDelegate common.Address) {

	fmt.Printf("--------------------\n")
	fmt.Printf("\n")

	// Main details
	fmt.Printf("Address:              %s\n", minipool.Address.Hex())
	if minipool.Penalties == 0 {
		fmt.Println("Penalties:            0")
	} else if minipool.Penalties < 3 {
		fmt.Printf("%sStrikes:              %d%s\n", colorYellow, minipool.Penalties, colorReset)
	} else {
		fmt.Printf("%sInfractions:          %d%s\n", colorRed, minipool.Penalties, colorReset)
	}
	fmt.Printf("Status updated:       %s\n", minipool.Status.StatusTime.Format(TimeFormat))
	fmt.Printf("Node fee:             %f%%\n", minipool.Node.Fee*100)
	fmt.Printf("Node deposit:         %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Node.DepositBalance), 6))

	// Queue position
	if minipool.Queue.Position != 0 {
		fmt.Printf("Queue position:       %d\n", minipool.Queue.Position)
	}

	// RP ETH deposit details - prelaunch & staking minipools
	if minipool.Status.Status == types.Prelaunch || minipool.Status.Status == types.Staking {
		if minipool.User.DepositAssigned {
			fmt.Printf("RP ETH assigned:      %s\n", minipool.User.DepositAssignedTime.Format(TimeFormat))
			fmt.Printf("RP deposit:           %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.User.DepositBalance), 6))
		} else {
			fmt.Printf("RP ETH assigned:      no\n")
		}
	}

	// Validator details - prelaunch and staking minipools
	if minipool.Status.Status == types.Prelaunch ||
		minipool.Status.Status == types.Staking {
		fmt.Printf("Validator pubkey:     %s\n", hex.AddPrefix(minipool.ValidatorPubkey.Hex()))
		fmt.Printf("Validator index:      %d\n", minipool.Validator.Index)
		if minipool.Validator.Exists {
			if minipool.Validator.Active {
				fmt.Printf("Validator active:     yes\n")
			} else {
				fmt.Printf("Validator active:     no\n")
			}
			fmt.Printf("Validator balance:    %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Validator.Balance), 6))
			fmt.Printf("Expected rewards:     %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Validator.NodeBalance), 6))
		} else {
			fmt.Printf("Validator seen:       no\n")
		}
	}

	// Withdrawal details - withdrawable minipools
	if minipool.Status.Status == types.Withdrawable {
		fmt.Printf("Withdrawal available: yes\n")
	}

	// Delegate details
	if minipool.UseLatestDelegate {
		fmt.Printf("Use latest delegate:  yes\n")
	} else {
		fmt.Printf("Use latest delegate:  no\n")
	}
	fmt.Printf("Delegate address:     %s\n", cliutils.GetPrettyAddress(minipool.Delegate))
	fmt.Printf("Rollback delegate:    %s\n", cliutils.GetPrettyAddress(minipool.PreviousDelegate))
	fmt.Printf("Effective delegate:   %s\n", cliutils.GetPrettyAddress(minipool.EffectiveDelegate))

	if minipool.EffectiveDelegate != latestDelegate {
		fmt.Printf("%s*Minipool can be upgraded to delegate %s!%s\n", colorYellow, latestDelegate.Hex(), colorReset)
	}

	fmt.Printf("\n")

}
