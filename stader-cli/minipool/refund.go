package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/stader-labs/stader-node/shared/services/gas"
	"github.com/stader-labs/stader-node/shared/services/stader"
	"github.com/stader-labs/stader-node/shared/types/api"
	cliutils "github.com/stader-labs/stader-node/shared/utils/cli"
	"github.com/stader-labs/stader-node/shared/utils/math"
)

func refundMinipools(c *cli.Context) error {

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

	// Get refundable minipools
	refundableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		if minipool.RefundAvailable {
			refundableMinipools = append(refundableMinipools, minipool)
		}
	}

	// Check for refundable minipools
	if len(refundableMinipools) == 0 {
		fmt.Println("No minipools have refunds available.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(refundableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range refundableMinipools {
			options[mi+1] = fmt.Sprintf("%s (%.6f ETH to claim)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.RefundBalance), 6))
		}
		selected, _ := cliutils.Select("Please select a minipool to refund ETH from:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = refundableMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{refundableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = refundableMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range refundableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available for refund.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		canResponse, err := staderClient.CanRefundMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for refund transaction (%s)", err)
			break
		} else {
			gasInfo = canResponse.GasInfo
			totalGas += canResponse.GasInfo.EstGasLimit
			totalSafeGas += canResponse.GasInfo.SafeGasLimit
		}
	}
	gasInfo.EstGasLimit = totalGas
	gasInfo.SafeGasLimit = totalSafeGas

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasInfo, staderClient, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to refund %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Refund minipools
	for _, minipool := range selectedMinipools {
		response, err := staderClient.RefundMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not refund ETH from minipool %s: %s.\n", minipool.Address.Hex(), err)
			continue
		}

		fmt.Printf("Refunding minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(staderClient, response.TxHash)
		if _, err = staderClient.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not refund ETH from minipool %s: %s.\n", minipool.Address.Hex(), err)
		} else {
			fmt.Printf("Successfully refunded ETH from minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
