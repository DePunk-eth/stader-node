package config

import cfgtypes "github.com/stader-labs/stader-node/shared/types/config"

func createNativeMetricsStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "Would you like to enable the daemon's metrics feature? This will allow you to access the Stader network's metrics and the metrics for your own node wallet in the Grafana dashboard."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		if wiz.md.Config.EnableMetrics.Value == false {
			modal.focus(0)
		} else {
			modal.focus(1)
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			wiz.md.Config.EnableMetrics.Value = true
		} else {
			wiz.md.Config.EnableMetrics.Value = false
		}
		if wiz.md.Config.Smartnode.Network.Value.(cfgtypes.Network) == cfgtypes.Network_Zhejiang {
			wiz.nativeFinishedModal.show()
		} else {
			wiz.nativeMevModal.show()
		}
	}

	back := func() {
		wiz.nativeUseFallbackModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"No", "Yes"},
		[]string{},
		76,
		"Metrics",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-native-metrics",
	)

}
