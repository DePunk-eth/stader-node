package collectors

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/stader-labs/stader-node/shared/services/config"
	"github.com/stader-labs/stader-node/stader/api/node"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for Snapshot metrics
type SnapshotCollector struct {
	// the number of active Snashot proposals
	activeProposals *prometheus.Desc

	// the number of past Snapshot proposals
	closedProposals *prometheus.Desc

	// the number of votes on active Snapshot proposals
	votesActiveProposals *prometheus.Desc

	// the number of votes on closed Snapshot proposals
	votesClosedProposals *prometheus.Desc

	// The current node voting power on Snapshot
	nodeVotingPower *prometheus.Desc

	// The current delegate voting power on Snapshot
	delegateVotingPower *prometheus.Desc

	// The Rocket Pool config
	cfg *config.StaderConfig

	// the node wallet address
	nodeAddress common.Address

	// the delegate address
	delegateAddress common.Address
}

// Create a new SnapshotCollector instance
func NewSnapshotCollector(rp *rocketpool.RocketPool, cfg *config.StaderConfig, nodeAddress common.Address, delegateAddres common.Address) *SnapshotCollector {
	subsystem := "snapshot"
	return &SnapshotCollector{
		activeProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_active"),
			"The number of active Snapshot proposals",
			nil, nil,
		),
		closedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_closed"),
			"The number of closed Snapshot proposals",
			nil, nil,
		),
		votesActiveProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_active"),
			"The number of votes from user/delegate on active Snapshot proposals",
			nil, nil,
		),
		votesClosedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_closed"),
			"The number of votes from user/delegate on closed Snapshot proposals",
			nil, nil,
		),
		nodeVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "node_vp"),
			"The node current voting power on Snapshot",
			nil, nil,
		),
		delegateVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "delegate_vp"),
			"The delegate current voting power on Snapshot",
			nil, nil,
		),
		cfg:             cfg,
		nodeAddress:     nodeAddress,
		delegateAddress: delegateAddres,
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SnapshotCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.activeProposals
	channel <- collector.closedProposals
	channel <- collector.votesActiveProposals
	channel <- collector.votesClosedProposals
	channel <- collector.nodeVotingPower
	channel <- collector.delegateVotingPower
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SnapshotCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	activeProposals := float64(0)
	closedProposals := float64(0)
	votesActiveProposals := float64(0)
	votesClosedProposals := float64(0)
	handledProposals := map[string]bool{}
	nodeVotingPower := float64(0)
	delegateVotingPower := float64(0)

	// Get the number of votes on Snapshot proposals
	wg.Go(func() error {
		votedProposals, err := node.GetSnapshotVotedProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.nodeAddress, collector.delegateAddress)
		if err != nil {
			return fmt.Errorf("Error getting Snapshot voted proposals: %w", err)
		}

		for _, votedProposal := range votedProposals.Data.Votes {
			_, exists := handledProposals[votedProposal.Proposal.Id]
			if !exists {
				if votedProposal.Proposal.State == "active" {
					votesActiveProposals += 1
				} else {
					votesClosedProposals += 1
				}
				handledProposals[votedProposal.Proposal.Id] = true
			}
		}

		return nil
	})

	// Get the number of live Snapshot proposals
	wg.Go(func() error {
		proposals, err := node.GetSnapshotProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), "")
		if err != nil {
			return fmt.Errorf("Error getting Snapshot voted proposals: %w", err)
		}

		for _, proposal := range proposals.Data.Proposals {
			if proposal.State == "active" {
				activeProposals += 1
			} else {
				closedProposals += 1
			}
		}

		return nil
	})

	// Get the node's voting power
	wg.Go(func() error {
		votingPowerResponse, err := node.GetSnapshotVotingPower(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.nodeAddress)
		if err != nil {
			return fmt.Errorf("Error getting Snapshot voted proposals for node address: %w", err)
		}

		nodeVotingPower = votingPowerResponse.Data.Vp.Vp

		return nil
	})

	// Get the delegate's voting power
	wg.Go(func() error {
		votingPowerResponse, err := node.GetSnapshotVotingPower(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.delegateAddress)
		if err != nil {
			return fmt.Errorf("Error getting Snapshot voted proposals for delegate address: %w", err)
		}

		delegateVotingPower = votingPowerResponse.Data.Vp.Vp

		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	channel <- prometheus.MustNewConstMetric(
		collector.votesActiveProposals, prometheus.GaugeValue, votesActiveProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.votesClosedProposals, prometheus.GaugeValue, votesClosedProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.activeProposals, prometheus.GaugeValue, activeProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.closedProposals, prometheus.GaugeValue, closedProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.nodeVotingPower, prometheus.GaugeValue, nodeVotingPower)
	channel <- prometheus.MustNewConstMetric(
		collector.delegateVotingPower, prometheus.GaugeValue, delegateVotingPower)
}
