package node

import (
	"crypto/rand"
	"fmt"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/stader-labs/stader-minipool-go/utils/eth"
	"github.com/stader-labs/stader-node/shared/services/gas"
	"github.com/urfave/cli"
	"math/big"

	"github.com/stader-labs/stader-node/shared/services/stader"
	cliutils "github.com/stader-labs/stader-node/shared/utils/cli"
	"github.com/stader-labs/stader-node/shared/utils/math"
)

// Config
const DefaultMaxNodeFeeSlippage = 0.01 // 1% below current network fee

func nodeDeposit(c *cli.Context) error {

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

	//// Make sure ETH2 is on the correct chain
	//depositContractInfo, err := staderClient.DepositContractInfo()
	//if err != nil {
	//	return err
	//}
	//if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
	//	depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
	//	cliutils.PrintDepositMismatchError(
	//		depositContractInfo.RPNetwork,
	//		depositContractInfo.BeaconNetwork,
	//		depositContractInfo.RPDepositContract,
	//		depositContractInfo.BeaconDepositContract)
	//	return nil
	//}

	fmt.Println("Your eth2 client is on the correct network.\n")

	// Check if the fee distributor has been initialized
	//isInitializedResponse, err := staderClient.IsFeeDistributorInitialized()
	//if err != nil {
	//	return err
	//}
	//if !isInitializedResponse.IsInitialized {
	//	fmt.Println("Your fee distributor has not been initialized yet so you cannot create a new minipool.\nPlease run `rocketpool node initialize-fee-distributor` to initialize it first.")
	//	return nil
	//}

	// Post a warning about fee distribution
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("%sNOTE: by creating a new minipool, your node will automatically claim and distribute any balance you have in your fee distributor contract. If you don't want to claim your balance at this time, you should not create a new minipool.%s\nWould you like to continue?", colorYellow, colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get deposit amount
	/*
		var amount float64
		if c.String("amount") != "" {

			// Parse amount
			depositAmount, err := strconv.ParseFloat(c.String("amount"), 64)
			if err != nil {
				return fmt.Errorf("Invalid deposit amount '%s': %w", c.String("amount"), err)
			}
			amount = depositAmount

		} else {

			// Get deposit amount options
			amountOptions := []string{
				"32 ETH (minipool begins staking immediately)",
				"16 ETH (minipool begins staking after ETH is assigned)",
			}

			// Prompt for amount
			selected, _ := cliutils.Select("Please choose an amount of ETH to deposit:", amountOptions)
			switch selected {
			case 0:
				amount = 32
			case 1:
				amount = 16
			}

			// Get node status
			status, err := rp.NodeStatus()
			if err != nil {
				return err
			}

			// Get deposit amount options
			amountOptions := []string{
				"32 ETH (minipool begins staking immediately)",
				"16 ETH (minipool begins staking after ETH is assigned)",
			}
			if status.Trusted {
				amountOptions = append(amountOptions, "0 ETH  (minipool begins staking after ETH is assigned)")
			}

			// Prompt for amount
			selected, _ := cliutils.Select("Please choose an amount of ETH to deposit:", amountOptions)
			switch selected {
			case 0:
				amount = 32
			case 1:
				amount = 16
			case 2:
				amount = 0
			}

		}
	*/

	// Force 4 ETH minipools as the only option after much community discussion
	amountWei := eth.EthToWei(4.0)

	//// Get network node fees
	//nodeFees, err := staderClient.NodeFee()
	//if err != nil {
	//	return err
	//}
	//
	//// Get minimum node fee
	//var minNodeFee float64
	//if c.String("max-slippage") == "auto" {
	//
	//	// Use default max slippage
	//	minNodeFee = nodeFees.NodeFee - DefaultMaxNodeFeeSlippage
	//	if minNodeFee < nodeFees.MinNodeFee {
	//		minNodeFee = nodeFees.MinNodeFee
	//	}
	//
	//} else if c.String("max-slippage") != "" {
	//
	//	// Parse max slippage
	//	maxNodeFeeSlippagePerc, err := strconv.ParseFloat(c.String("max-slippage"), 64)
	//	if err != nil {
	//		return fmt.Errorf("Invalid maximum commission rate slippage '%s': %w", c.String("max-slippage"), err)
	//	}
	//	maxNodeFeeSlippage := maxNodeFeeSlippagePerc / 100
	//
	//	// Calculate min node fee
	//	minNodeFee = nodeFees.NodeFee - maxNodeFeeSlippage
	//	if minNodeFee < nodeFees.MinNodeFee {
	//		minNodeFee = nodeFees.MinNodeFee
	//	}
	//
	//} else {
	//
	//	// Prompt for min node fee
	//	if nodeFees.MinNodeFee == nodeFees.MaxNodeFee {
	//		fmt.Printf("Your minipool will use the current fixed commission rate of %.2f%%.\n", nodeFees.MinNodeFee*100)
	//		minNodeFee = nodeFees.MinNodeFee
	//	} else {
	//		minNodeFee = promptMinNodeFee(nodeFees.NodeFee, nodeFees.MinNodeFee)
	//	}
	//
	//}

	// Get minipool salt
	var salt *big.Int
	if c.String("salt") != "" {
		var success bool
		salt, success = big.NewInt(0).SetString(c.String("salt"), 0)
		if !success {
			return fmt.Errorf("Invalid minipool salt: %s", c.String("salt"))
		}
	} else {
		buffer := make([]byte, 32)
		_, err = rand.Read(buffer)
		if err != nil {
			return fmt.Errorf("Error generating random salt: %w", err)
		}
		salt = big.NewInt(0).SetBytes(buffer)
	}

	// Check deposit can be made
	//canDeposit, err := staderClient.CanNodeDeposit(amountWei, minNodeFee, salt)
	//if err != nil {
	//	return err
	//}
	//if !canDeposit.CanDeposit {
	//	fmt.Println("Cannot make node deposit:")
	//	if canDeposit.InsufficientBalance {
	//		fmt.Println("The node's ETH balance is insufficient.")
	//	}
	//	if canDeposit.InsufficientRplStake {
	//		fmt.Println("The node has not staked enough RPL to collateralize a new minipool.")
	//	}
	//	if canDeposit.InvalidAmount {
	//		fmt.Println("The deposit amount is invalid.")
	//	}
	//	if canDeposit.UnbondedMinipoolsAtMax {
	//		fmt.Println("The node cannot create any more unbonded minipools.")
	//	}
	//	if canDeposit.DepositDisabled {
	//		fmt.Println("Node deposits are currently disabled.")
	//	}
	//	if !canDeposit.InConsensus {
	//		fmt.Println("The RPL price and total effective staked RPL of the network are still being voted on by the Oracle DAO.\nPlease try again in a few minutes.")
	//	}
	//	return nil
	//}

	//if c.String("salt") != "" {
	//	fmt.Printf("Using custom salt %s, your minipool address will be %s.\n\n", c.String("salt"), canDeposit.MinipoolAddress.Hex())
	//}

	// Check to see if eth2 is synced
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorYellow := "\033[33m"
	syncResponse, err := staderClient.NodeSync()
	if err != nil {
		fmt.Printf("%s**WARNING**: Can't verify the sync status of your consensus client.\nYOU WILL LOSE ETH if your minipool is activated before it is fully synced.\n"+
			"Reason: %s\n%s", colorRed, err, colorReset)
	} else {
		if syncResponse.BcStatus.PrimaryClientStatus.IsSynced {
			fmt.Printf("Your consensus client is synced, you may safely create a minipool.\n")
		} else if syncResponse.BcStatus.FallbackEnabled {
			if syncResponse.BcStatus.FallbackClientStatus.IsSynced {
				fmt.Printf("Your fallback consensus client is synced, you may safely create a minipool.\n")
			} else {
				fmt.Printf("%s**WARNING**: neither your primary nor fallback consensus clients are fully synced.\nYOU WILL LOSE ETH if your minipool is activated before they are fully synced.\n%s", colorRed, colorReset)
			}
		} else {
			fmt.Printf("%s**WARNING**: your primary consensus client is either not fully synced or offline and you do not have a fallback client configured.\nYOU WILL LOSE ETH if your minipool is activated before it is fully synced.\n%s", colorRed, colorReset)
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(rocketpool.GasInfo{
		EstGasLimit:  10000000,
		SafeGasLimit: 50000000,
	}, staderClient, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf(
		"You are about to deposit %.6f ETH to create a minipool with a minimum possible commission rate of %f%%.\n"+
			"%sARE YOU SURE YOU WANT TO DO THIS? Running a minipool is a long-term commitment, and this action cannot be undone!%s",
		math.RoundDown(eth.WeiToEth(amountWei), 6),
		0.0,
		colorYellow,
		colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	operatorName := c.String("operator-name")
	// TODO: validate this address
	operatorRewardAddress := c.String("operator-rewarder-address")
	// Make deposit
	response, err := staderClient.NodeDeposit(amountWei, salt, operatorName, operatorRewardAddress, true)
	if err != nil {
		return err
	}

	// Log and wait for the minipool address
	fmt.Printf("Creating minipool...\n")
	cliutils.PrintTransactionHash(staderClient, response.TxHash)
	_, err = staderClient.WaitForTransaction(response.TxHash)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node deposit of %.6f ETH was made successfully!\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	fmt.Printf("Your new minipool's address is: %s\n", response.MinipoolAddress)
	fmt.Printf("The validator pubkey is: %s\n\n", response.ValidatorPubkey.Hex())

	fmt.Println("Your minipool is now in Initialized status.")
	fmt.Println("Once the 4 ETH deposit has been matched by the staking pool, it will move to Prelaunch status.")
	fmt.Printf("After that, it will move to Staking status once %s have passed.\n", response.ScrubPeriod)
	fmt.Println("You can watch its progress using `rocketpool service logs node`.")

	return nil

}
