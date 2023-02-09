package watchtower

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/stader-labs/stader-node/shared/services"
	"github.com/stader-labs/stader-node/shared/services/beacon"
	"github.com/stader-labs/stader-node/shared/services/config"
	"github.com/stader-labs/stader-node/shared/services/contracts"
	"github.com/stader-labs/stader-node/shared/services/wallet"
	"github.com/stader-labs/stader-node/shared/utils/api"
	"github.com/stader-labs/stader-node/shared/utils/eth1"
	"github.com/stader-labs/stader-node/shared/utils/log"
	mathutils "github.com/stader-labs/stader-node/shared/utils/math"
)

const MessengerAbi = `[
    {
      "inputs": [],
      "name": "rateStale",
      "outputs": [
        {
          "internalType": "bool",
          "name": "",
          "type": "bool"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [],
      "name": "submitRate",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ]`

// Settings
const BlocksPerTurn = 75 // Approx. 15 minutes

// Submit RPL price task
type submitRplPrice struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.StaderConfig
	ec  rocketpool.ExecutionClient
	w   *wallet.Wallet
	rp  *rocketpool.RocketPool
	oio *contracts.OneInchOracle
	bc  beacon.Client
}

// Create submit RPL price task
func newSubmitRplPrice(c *cli.Context, logger log.ColorLogger) (*submitRplPrice, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	oio, err := services.GetOneInchOracle(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	return &submitRplPrice{
		c:   c,
		log: logger,
		cfg: cfg,
		ec:  ec,
		w:   w,
		rp:  rp,
		oio: oio,
		bc:  bc,
	}, nil

}

// Submit RPL price
func (t *submitRplPrice) run() error {

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Data
	var wg errgroup.Group
	var nodeTrusted bool
	var submitPricesEnabled bool

	// Get data
	wg.Go(func() error {
		var err error
		nodeTrusted, err = trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		submitPricesEnabled, err = protocol.GetSubmitPricesEnabled(t.rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return err
	}

	// Check node trusted status & settings
	if !(nodeTrusted && submitPricesEnabled) {
		return nil
	}

	// Check if Optimism rate is stale and submit
	err = t.submitOptimismPrice()
	if err != nil {
		// Error is not fatal for this task so print and continue
		t.log.Printf("Error submitting Optimism price: %q\n", err)
	}

	// Log
	t.log.Println("Checking for RPL price checkpoint...")

	// Get block to submit price for
	blockNumber, err := t.getLatestReportableBlock()
	if err != nil {
		return err
	}

	// Check if a submission needs to be made
	pricesBlock, err := network.GetPricesBlock(t.rp, nil)
	if err != nil {
		return err
	}
	if blockNumber <= pricesBlock {
		return nil
	}

	// Get the time of the block
	header, err := t.ec.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(blockNumber))
	if err != nil {
		return err
	}
	blockTime := time.Unix(int64(header.Time), 0)

	// Get the Beacon block corresponding to this time
	eth2Config, err := t.bc.GetEth2Config()
	if err != nil {
		return err
	}
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	timeSinceGenesis := blockTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot

	// Check if the epoch is finalized yet
	epoch := slotNumber / eth2Config.SlotsPerEpoch
	beaconHead, err := t.bc.GetBeaconHead()
	if err != nil {
		return err
	}
	finalizedEpoch := beaconHead.FinalizedEpoch
	if epoch > finalizedEpoch {
		t.log.Printlnf("Prices must be reported for EL block %d, waiting until Epoch %d is finalized (currently %d)", blockNumber, epoch, finalizedEpoch)
		return nil
	}

	// Log
	t.log.Printlnf("Getting RPL price for block %d...", blockNumber)

	// Get RPL price at block
	rplPrice, err := t.getRplPrice(blockNumber)
	if err != nil {
		return err
	}

	// Calculate the total effective RPL stake on the network
	zero := new(big.Int).SetUint64(0)
	effectiveRplStake, err := node.CalculateTotalEffectiveRPLStake(t.rp, zero, zero, rplPrice, nil)
	if err != nil {
		return fmt.Errorf("Error getting total effective RPL stake: %w", err)
	}

	// Log
	t.log.Printlnf("RPL price: %.6f ETH", mathutils.RoundDown(eth.WeiToEth(rplPrice), 6))

	// Check if we have reported these specific values before
	hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockPrices(nodeAccount.Address, blockNumber, rplPrice, effectiveRplStake)
	if err != nil {
		return err
	}
	if hasSubmittedSpecific {
		return nil
	}

	// We haven't submitted these values, check if we've submitted any for this block so we can log it
	hasSubmitted, err := t.hasSubmittedBlockPrices(nodeAccount.Address, blockNumber)
	if err != nil {
		return err
	}
	if hasSubmitted {
		t.log.Printlnf("Have previously submitted out-of-date prices for block %d, trying again...", blockNumber)
	}

	// Log
	t.log.Println("Submitting RPL price...")

	// Submit RPL price
	if err := t.submitRplPrice(blockNumber, rplPrice, effectiveRplStake); err != nil {
		return fmt.Errorf("Could not submit RPL price: %w", err)
	}

	// Return
	return nil

}

// Get the latest block number to report RPL price for
func (t *submitRplPrice) getLatestReportableBlock() (uint64, error) {

	// Require eth client synced
	if err := services.RequireEthClientSynced(t.c); err != nil {
		return 0, err
	}

	latestBlock, err := network.GetLatestReportablePricesBlock(t.rp, nil)
	if err != nil {
		return 0, fmt.Errorf("Error getting latest reportable block: %w", err)
	}
	return latestBlock.Uint64(), nil

}

// Check whether prices for a block has already been submitted by the node
func (t *submitRplPrice) hasSubmittedBlockPrices(nodeAddress common.Address, blockNumber uint64) (bool, error) {

	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.prices.submitted.node"), nodeAddress.Bytes(), blockNumberBuf))

}

// Check whether specific prices for a block has already been submitted by the node
func (t *submitRplPrice) hasSubmittedSpecificBlockPrices(nodeAddress common.Address, blockNumber uint64, rplPrice, effectiveRplStake *big.Int) (bool, error) {

	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)

	rplPriceBuf := make([]byte, 32)
	rplPrice.FillBytes(rplPriceBuf)

	effectiveRplStakeBuf := make([]byte, 32)
	effectiveRplStake.FillBytes(effectiveRplStakeBuf)

	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.prices.submitted.node"), nodeAddress.Bytes(), blockNumberBuf, rplPriceBuf, effectiveRplStakeBuf))

}

// Get RPL price at block
func (t *submitRplPrice) getRplPrice(blockNumber uint64) (*big.Int, error) {

	// Require 1inch oracle contract
	if err := services.RequireOneInchOracle(t.c); err != nil {
		return nil, err
	}

	// Get RPL token address
	rplAddress := common.HexToAddress(t.cfg.Smartnode.GetRplTokenAddress())

	// Initialize call options
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(int64(blockNumber)),
	}

	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, opts.BlockNumber)
	if err != nil {
		return nil, err
	}

	// Generate an OIO wrapper using the client
	oio, err := contracts.NewOneInchOracle(common.HexToAddress(t.cfg.Smartnode.GetOneInchOracleAddress()), client.Client)
	if err != nil {
		return nil, err
	}

	// Get RPL price
	rplPrice, err := oio.GetRateToEth(opts, rplAddress, true)
	if err != nil {
		return nil, fmt.Errorf("Could not get RPL price at block %d: %w", blockNumber, err)
	}

	// Return
	return rplPrice, nil

}

func (t *submitRplPrice) printMessage(message string) {
	t.log.Println(message)
}

// Submit RPL price and total effective RPL stake
func (t *submitRplPrice) submitRplPrice(blockNumber uint64, rplPrice, effectiveRplStake *big.Int) error {

	// Log
	t.log.Printlnf("Submitting RPL price for block %d...", blockNumber)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := network.EstimateSubmitPricesGas(t.rp, blockNumber, rplPrice, effectiveRplStake, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit RPL price: %w", err)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(WatchtowerMaxFee)
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(WatchtowerMaxPriorityFee)
	opts.GasLimit = gasInfo.SafeGasLimit

	// Submit RPL price
	hash, err := network.SubmitPrices(t.rp, blockNumber, rplPrice, effectiveRplStake, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted RPL price for block %d.", blockNumber)

	// Return
	return nil

}

// Checks if Optimism rate is stale and if it's our turn to submit, calls submitRate on the messenger
func (t *submitRplPrice) submitOptimismPrice() error {
	priceMessengerAddress := t.cfg.Smartnode.GetOptimismMessengerAddress()

	if priceMessengerAddress == "" {
		// No price messenger deployed on the current network
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return fmt.Errorf("Failed getting transactor: %q", err)
	}

	// Construct the price messenger contract instance
	parsed, err := abi.JSON(strings.NewReader(MessengerAbi))
	if err != nil {
		return fmt.Errorf("Failed decoding ABI: %q", err)
	}

	addr := common.HexToAddress(priceMessengerAddress)
	priceMessengerContract := bind.NewBoundContract(addr, parsed, t.ec, t.ec, t.ec)
	priceMessenger := rocketpool.Contract{
		Contract: priceMessengerContract,
		Address:  &addr,
		ABI:      &parsed,
		Client:   t.ec,
	}

	// Check if the rate is stale
	var out []interface{}
	err = priceMessengerContract.Call(nil, &out, "rateStale")

	if err != nil {
		return fmt.Errorf("Failed to query rate staleness: %q", err)
	}

	rateStale := *abi.ConvertType(out[0], new(bool)).(*bool)

	if !rateStale {
		// Nothing to do
		return nil
	}

	// Get total number of ODAO members
	count, err := trustednode.GetMemberCount(t.rp, nil)
	if err != nil {
		return fmt.Errorf("Failed to get member count: %q", err)
	}

	// Find out which index we are
	var index = uint64(0)
	for i := uint64(0); i < count; i++ {
		addr, err := trustednode.GetMemberAt(t.rp, i, nil)
		if err != nil {
			return fmt.Errorf("Failed to get member at %d: %q", i, err)
		}

		if bytes.Compare(addr.Bytes(), opts.From.Bytes()) == 0 {
			index = i
			break
		}
	}

	// Get current block number
	blockNumber, err := t.ec.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get block number: %q", err)
	}

	// Calculate whose turn it is to submit
	indexToSubmit := (blockNumber / BlocksPerTurn) % count

	if index == indexToSubmit {

		// Temporary gas calculations until this gets put into a binding
		input, err := priceMessenger.ABI.Pack("submitRate")
		if err != nil {
			return fmt.Errorf("Could not encode input data: %w", err)
		}

		// Estimate gas limit
		gasLimit, err := t.rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     opts.From,
			To:       priceMessenger.Address,
			GasPrice: big.NewInt(0), // use 0 gwei for simulation
			Value:    opts.Value,
			Data:     input,
		})
		if err != nil {
			return fmt.Errorf("Error estimating gas limit of submitOptimismPrice: %w", err)
		}

		// Get the safe gas limit
		safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
		if gasLimit > rocketpool.MaxGasLimit {
			gasLimit = rocketpool.MaxGasLimit
		}
		if safeGasLimit > rocketpool.MaxGasLimit {
			safeGasLimit = rocketpool.MaxGasLimit
		}
		gasInfo := rocketpool.GasInfo{
			EstGasLimit:  gasLimit,
			SafeGasLimit: safeGasLimit,
		}

		// Print the gas info
		maxFee := eth.GweiToWei(WatchtowerMaxFee)
		if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = eth.GweiToWei(WatchtowerMaxPriorityFee)
		opts.GasLimit = gasInfo.SafeGasLimit

		t.log.Println("Submitting rate to Optimism...")

		// Submit rates
		tx, err := priceMessenger.Transact(opts, "submitRate")
		if err != nil {
			return fmt.Errorf("Failed to submit rate: %q", err)
		}

		// Print TX info and wait for it to be included in a block
		err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, t.log)
		if err != nil {
			return err
		}

		// Log
		t.log.Printlnf("Successfully submitted Optimism price for block %d.", blockNumber)

	}

	return nil
}
