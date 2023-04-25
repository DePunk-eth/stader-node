package collector

import (
	"fmt"

	"github.com/stader-labs/stader-node/stader-lib/stader"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stader-labs/stader-node/shared/services/beacon"

	"github.com/prometheus/client_golang/prometheus"
)

// Represents the collector for the stader network metrics
type OperatorCollector struct {
	ActiveValidators                     *prometheus.Desc
	QueuedValidators                     *prometheus.Desc
	SlashedValidators                    *prometheus.Desc
	ExitingValidators                    *prometheus.Desc
	WithdrawnValidators                  *prometheus.Desc
	TotalSDBonded                        *prometheus.Desc
	UnclaimedClRewards                   *prometheus.Desc
	UnclaimedNonSocializingPoolElRewards *prometheus.Desc
	CumulativePenalty                    *prometheus.Desc
	UnclaimedSocializingPoolELRewards    *prometheus.Desc
	UnclaimedSocializingPoolSdRewards    *prometheus.Desc
	ClaimedSocializingPoolSdRewards      *prometheus.Desc
	ClaimedSocializingPoolElRewards      *prometheus.Desc
	NextRewardCycleTime                  *prometheus.Desc

	// The beacon client
	bc beacon.Client

	// The eth1 client
	ec stader.ExecutionClient

	// The node's address
	nodeAddress common.Address

	// The thread-safe locker for the network state
	stateLocker *StateCache

	// Prefix for logging
	logPrefix string
}

// Create a new NetworkCollector instance
func NewOperatorCollector(
	bc beacon.Client,
	ec stader.ExecutionClient,
	nodeAddress common.Address,
	stateLocker *StateCache,
) *OperatorCollector {
	return &OperatorCollector{
		ActiveValidators: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, ActiveValidators), "", nil, nil,
		),
		QueuedValidators: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, QueuedValidators), "", nil, nil,
		),
		SlashedValidators: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, SlashedValidators), "", nil, nil,
		),
		ExitingValidators: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, ExitingValidators), "", nil, nil,
		),
		WithdrawnValidators: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, WithdrawnValidators), "", nil, nil,
		),
		TotalSDBonded: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, TotalSDBonded), "", nil, nil,
		),
		UnclaimedClRewards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, UnclaimedCLRewards), "", nil, nil,
		),
		UnclaimedNonSocializingPoolElRewards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, UnclaimedSocializingPoolELRewards), "", nil, nil,
		),
		CumulativePenalty: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, CumulativePenalty), "", nil, nil),
		UnclaimedSocializingPoolELRewards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, UnclaimedSocializingPoolELRewards), "", nil, nil,
		),
		UnclaimedSocializingPoolSdRewards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, UnclaimedSocializingPoolSdRewards), "", nil, nil,
		),
		ClaimedSocializingPoolSdRewards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, ClaimedSocializingPoolSDrewards), "", nil, nil),
		ClaimedSocializingPoolElRewards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, ClaimedSocializingPoolELRewards), "", nil, nil),
		NextRewardCycleTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RewardSub, NextRewardCycleTime), "", nil, nil),

		bc:          bc,
		ec:          ec,
		nodeAddress: nodeAddress,
		stateLocker: stateLocker,
		logPrefix:   "Operator Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *OperatorCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.ActiveValidators
	channel <- collector.QueuedValidators
	channel <- collector.SlashedValidators
	channel <- collector.ExitingValidators
	channel <- collector.WithdrawnValidators
	channel <- collector.TotalSDBonded
	channel <- collector.UnclaimedClRewards
	channel <- collector.UnclaimedNonSocializingPoolElRewards
	channel <- collector.CumulativePenalty
	channel <- collector.UnclaimedSocializingPoolELRewards
	channel <- collector.UnclaimedSocializingPoolSdRewards
	channel <- collector.ClaimedSocializingPoolSdRewards
	channel <- collector.ClaimedSocializingPoolElRewards
	channel <- collector.NextRewardCycleTime
}

// Collect the latest metric values and pass them to Prometheus
func (collector *OperatorCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()

	channel <- prometheus.MustNewConstMetric(collector.ActiveValidators, prometheus.GaugeValue, float64(state.StaderNetworkDetails.ActiveValidators.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.QueuedValidators, prometheus.GaugeValue, float64(state.StaderNetworkDetails.QueuedValidators.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.SlashedValidators, prometheus.GaugeValue, float64(state.StaderNetworkDetails.SlashedValidators.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.ExitingValidators, prometheus.GaugeValue, float64(state.StaderNetworkDetails.ExitingValidators.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.WithdrawnValidators, prometheus.GaugeValue, float64(state.StaderNetworkDetails.WithdrawnValidators.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.TotalSDBonded, prometheus.GaugeValue, float64(state.StaderNetworkDetails.TotalStakedSd.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.UnclaimedClRewards, prometheus.GaugeValue, float64(state.StaderNetworkDetails.UnclaimedClRewards.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.UnclaimedNonSocializingPoolElRewards, prometheus.GaugeValue, float64(state.StaderNetworkDetails.UnclaimedNonSocializingPoolElRewards.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.CumulativePenalty, prometheus.GaugeValue, float64(state.StaderNetworkDetails.CumulativePenalty.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.UnclaimedSocializingPoolELRewards, prometheus.GaugeValue, float64(state.StaderNetworkDetails.UnclaimedSocializingPoolElRewards.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.UnclaimedSocializingPoolSdRewards, prometheus.GaugeValue, float64(state.StaderNetworkDetails.UnclaimedSocializingPoolSDRewards.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.ClaimedSocializingPoolSdRewards, prometheus.GaugeValue, float64(state.StaderNetworkDetails.ClaimedSocializingPoolSdRewards.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.ClaimedSocializingPoolElRewards, prometheus.GaugeValue, float64(state.StaderNetworkDetails.ClaimedSocializingPoolElRewards.Int64()))
	channel <- prometheus.MustNewConstMetric(collector.NextRewardCycleTime, prometheus.GaugeValue, float64(10))

}

// Log error messages
func (collector *OperatorCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}