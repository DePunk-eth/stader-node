package rewards

import (
	"fmt"

	cfgtypes "github.com/stader-labs/stader-node/shared/types/config"
)

type rewardsIntervalInfo struct {
	rewardsRulesetVersion uint64
	mainnetStartInterval  uint64
	praterStartInterval   uint64
	generator             treeGeneratorImpl
}

func (r *rewardsIntervalInfo) GetStartInterval(network cfgtypes.Network) (uint64, error) {
	switch network {
	case cfgtypes.Network_Mainnet:
		return r.mainnetStartInterval, nil
	case cfgtypes.Network_Prater:
		return r.praterStartInterval, nil
	case cfgtypes.Network_Devnet:
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown network: %s", string(network))
	}
}
