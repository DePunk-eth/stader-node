package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/stader-labs/stader-node/shared/services/gas"
	"github.com/stader-labs/stader-node/shared/services/stader"
	cliutils "github.com/stader-labs/stader-node/shared/utils/cli"
	"github.com/stader-labs/stader-node/shared/utils/math"
)

func nodeStakeRpl(c *cli.Context) error {

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

	// Get node status
	status, err := staderClient.NodeStatus()
	if err != nil {
		return err
	}

	// If a custom nonce is set, print the multi-transaction warning
	if c.GlobalUint64("nonce") != 0 {
		cliutils.PrintMultiTransactionNonceWarning()
	}

	// Check for fixed-supply RPL balance
	rplBalance := *(status.AccountBalances.RPL)
	if status.AccountBalances.FixedSupplyRPL.Cmp(big.NewInt(0)) > 0 {

		// Confirm swapping RPL
		if c.Bool("swap") || cliutils.Confirm(fmt.Sprintf("The node has a balance of %.6f old RPL. Would you like to swap it for new RPL before staking?", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))) {

			// Check allowance
			allowance, err := staderClient.GetNodeSwapRplAllowance()
			if err != nil {
				return err
			}

			if allowance.Allowance.Cmp(status.AccountBalances.FixedSupplyRPL) < 0 {
				fmt.Println("Before swapping legacy RPL for new RPL, you must first give the new RPL contract approval to interact with your legacy RPL.")
				fmt.Println("This only needs to be done once for your node.")

				// If a custom nonce is set, print the multi-transaction warning
				if c.GlobalUint64("nonce") != 0 {
					cliutils.PrintMultiTransactionNonceWarning()
				}

				// Calculate max uint256 value
				maxApproval := big.NewInt(2)
				maxApproval = maxApproval.Exp(maxApproval, big.NewInt(256), nil)
				maxApproval = maxApproval.Sub(maxApproval, big.NewInt(1))

				// Get approval gas
				approvalGas, err := staderClient.NodeSwapRplApprovalGas(maxApproval)
				if err != nil {
					return err
				}
				// Assign max fees
				err = gas.AssignMaxFeeAndLimit(approvalGas.GasInfo, staderClient, c.Bool("yes"))
				if err != nil {
					return err
				}

				// Prompt for confirmation
				if !(c.Bool("yes") || cliutils.Confirm("Do you want to let the new RPL contract interact with your legacy RPL?")) {
					fmt.Println("Cancelled.")
					return nil
				}

				// Approve RPL for swapping
				response, err := staderClient.NodeSwapRplApprove(maxApproval)
				if err != nil {
					return err
				}
				hash := response.ApproveTxHash
				fmt.Printf("Approving legacy RPL for swapping...\n")
				cliutils.PrintTransactionHash(staderClient, hash)
				if _, err = staderClient.WaitForTransaction(hash); err != nil {
					return err
				}
				fmt.Println("Successfully approved access to legacy RPL.")

				// If a custom nonce is set, increment it for the next transaction
				if c.GlobalUint64("nonce") != 0 {
					staderClient.IncrementCustomNonce()
				}
			}

			// Check RPL can be swapped
			canSwap, err := staderClient.CanNodeSwapRpl(status.AccountBalances.FixedSupplyRPL)
			if err != nil {
				return err
			}
			if !canSwap.CanSwap {
				fmt.Println("Cannot swap RPL:")
				if canSwap.InsufficientBalance {
					fmt.Println("The node's old RPL balance is insufficient.")
				}
				return nil
			}
			fmt.Println("RPL Swap Gas Info:")
			// Assign max fees
			err = gas.AssignMaxFeeAndLimit(canSwap.GasInfo, staderClient, c.Bool("yes"))
			if err != nil {
				return err
			}

			// Prompt for confirmation
			if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to swap %.6f old RPL for new RPL?", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6)))) {
				fmt.Println("Cancelled.")
				return nil
			}

			// Swap RPL
			swapResponse, err := staderClient.NodeSwapRpl(status.AccountBalances.FixedSupplyRPL)
			if err != nil {
				return err
			}

			fmt.Printf("Swapping old RPL for new RPL...\n")
			cliutils.PrintTransactionHash(staderClient, swapResponse.SwapTxHash)
			if _, err = staderClient.WaitForTransaction(swapResponse.SwapTxHash); err != nil {
				return err
			}

			// Log
			fmt.Printf("Successfully swapped %.6f old RPL for new RPL.\n", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))
			fmt.Println("")

			// If a custom nonce is set, increment it for the next transaction
			if c.GlobalUint64("nonce") != 0 {
				staderClient.IncrementCustomNonce()
			}

			// Get new account RPL balance
			rplBalance.Add(status.AccountBalances.RPL, status.AccountBalances.FixedSupplyRPL)

		}

	}

	// Get stake mount
	var amountWei *big.Int
	if c.String("amount") == "min" {

		// Set amount to min per minipool RPL stake
		rplPrice, err := staderClient.RplPrice()
		if err != nil {
			return err
		}
		amountWei = rplPrice.MinPerMinipoolRplStake

	} else if c.String("amount") == "max" {

		// Set amount to max per minipool RPL stake
		rplPrice, err := staderClient.RplPrice()
		if err != nil {
			return err
		}
		amountWei = rplPrice.MaxPerMinipoolRplStake

	} else if c.String("amount") == "all" {

		// Set amount to node's entire RPL balance
		amountWei = &rplBalance

	} else if c.String("amount") != "" {

		// Parse amount
		stakeAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid stake amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(stakeAmount)

	} else {

		// Get min/max per minipool RPL stake amounts
		rplPrice, err := staderClient.RplPrice()
		if err != nil {
			return err
		}
		minAmount := rplPrice.MinPerMinipoolRplStake
		maxAmount := rplPrice.MaxPerMinipoolRplStake

		// Prompt for amount option
		amountOptions := []string{
			fmt.Sprintf("The minimum minipool stake amount (%.6f RPL)?", math.RoundUp(eth.WeiToEth(minAmount), 6)),
			fmt.Sprintf("The maximum effective minipool stake amount (%.6f RPL)?", math.RoundUp(eth.WeiToEth(maxAmount), 6)),
			fmt.Sprintf("Your entire RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(&rplBalance), 6)),
			"A custom amount",
		}
		selected, _ := cliutils.Select("Please choose an amount of RPL to stake:", amountOptions)
		switch selected {
		case 0:
			amountWei = minAmount
		case 1:
			amountWei = maxAmount
		case 2:
			amountWei = &rplBalance
		}

		// Prompt for custom amount
		if amountWei == nil {
			inputAmount := cliutils.Prompt("Please enter an amount of RPL to stake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			stakeAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid stake amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(stakeAmount)
		}

	}

	// Check allowance
	allowance, err := staderClient.GetNodeStakeRplAllowance()
	if err != nil {
		return err
	}

	if allowance.Allowance.Cmp(amountWei) < 0 {
		fmt.Println("Before staking RPL, you must first give the staking contract approval to interact with your RPL.")
		fmt.Println("This only needs to be done once for your node.")

		// If a custom nonce is set, print the multi-transaction warning
		if c.GlobalUint64("nonce") != 0 {
			cliutils.PrintMultiTransactionNonceWarning()
		}

		// Calculate max uint256 value
		maxApproval := big.NewInt(2)
		maxApproval = maxApproval.Exp(maxApproval, big.NewInt(256), nil)
		maxApproval = maxApproval.Sub(maxApproval, big.NewInt(1))

		// Get approval gas
		approvalGas, err := staderClient.NodeStakeRplApprovalGas(maxApproval)
		if err != nil {
			return err
		}
		// Assign max fees
		err = gas.AssignMaxFeeAndLimit(approvalGas.GasInfo, staderClient, c.Bool("yes"))
		if err != nil {
			return err
		}

		// Prompt for confirmation
		if !(c.Bool("yes") || cliutils.Confirm("Do you want to let the staking contract interact with your RPL?")) {
			fmt.Println("Cancelled.")
			return nil
		}

		// Approve RPL for staking
		response, err := staderClient.NodeStakeRplApprove(maxApproval)
		if err != nil {
			return err
		}
		hash := response.ApproveTxHash
		fmt.Printf("Approving RPL for staking...\n")
		cliutils.PrintTransactionHash(staderClient, hash)
		if _, err = staderClient.WaitForTransaction(hash); err != nil {
			return err
		}
		fmt.Println("Successfully approved staking access to RPL.")

		// If a custom nonce is set, increment it for the next transaction
		if c.GlobalUint64("nonce") != 0 {
			staderClient.IncrementCustomNonce()
		}
	}

	// Check RPL can be staked
	canStake, err := staderClient.CanNodeStakeRpl(amountWei)
	if err != nil {
		return err
	}
	if !canStake.CanStake {
		fmt.Println("Cannot stake RPL:")
		if canStake.InsufficientBalance {
			fmt.Println("The node's RPL balance is insufficient.")
		}
		if !canStake.InConsensus {
			fmt.Println("The RPL price and total effective staked RPL of the network are still being voted on by the Oracle DAO.\nPlease try again in a few minutes.")
		}
		return nil
	}

	fmt.Println("RPL Stake Gas Info:")
	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canStake.GasInfo, staderClient, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to stake %.6f RPL? You will not be able to unstake this RPL until you exit your validators and close your minipools, or reach over 150%% collateral!", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stake RPL
	stakeResponse, err := staderClient.NodeStakeRpl(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Staking RPL...\n")
	cliutils.PrintTransactionHash(staderClient, stakeResponse.StakeTxHash)
	if _, err = staderClient.WaitForTransaction(stakeResponse.StakeTxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully staked %.6f RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
