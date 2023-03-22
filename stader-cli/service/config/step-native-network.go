package config

import (
	cfgtypes "github.com/stader-labs/stader-node/shared/types/config"
)

func createNativeNetworkStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	networks := wiz.md.Config.StaderNode.Network.Options
	networkNames := []string{}
	networkDescriptions := []string{}
	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
		networkDescriptions = append(networkDescriptions, network.Description)
	}

	helperText := "Let's start by choosing which network you'd like to use.\n\n"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.StaderNode.Network.Options {
			if option.Value == wiz.md.Config.StaderNode.Network.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		newNetwork := networks[buttonIndex].Value.(cfgtypes.Network)
		wiz.md.Config.ChangeNetwork(newNetwork)
		wiz.nativeEcModal.show()
	}

	back := func() {
		wiz.nativeWelcomeModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		networkNames,
		networkDescriptions,
		70,
		"Network",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-native-network",
	)

}
