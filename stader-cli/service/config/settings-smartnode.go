package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/stader-labs/stader-node/shared/services/config"
	cfgtypes "github.com/stader-labs/stader-node/shared/types/config"
)

// The page wrapper for the Stader Node config
type SmartnodeConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// Creates a new page for the Stader Node settings
func NewSmartnodeConfigPage(home *settingsHome) *SmartnodeConfigPage {

	configPage := &SmartnodeConfigPage{
		home: home,
	}

	configPage.createContent()
	configPage.page = newPage(
		home.homePage,
		"settings-smartnode",
		"Stader Node and TX Fees",
		"Select this to configure the settings for the Stader Node itself, including the defaults and limits on transaction fees.",
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *SmartnodeConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the Stader Node settings page
func (configPage *SmartnodeConfigPage) createContent() {

	// Create the layout
	masterConfig := configPage.home.md.Config
	layout := newStandardLayout()
	configPage.layout = layout
	layout.createForm(&masterConfig.Stadernode.Network, "Stader Node and TX Fee Settings")

	// Return to the home page after pressing Escape
	layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			// Close all dropdowns and break if one was open
			for _, param := range configPage.layout.parameters {
				dropDown, ok := param.item.(*DropDown)
				if ok && dropDown.open {
					dropDown.CloseList(configPage.home.md.app)
					return nil
				}
			}

			// Return to the home page
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	formItems := createParameterizedFormItems(masterConfig.Stadernode.GetParameters(), layout.descriptionBox)
	for _, formItem := range formItems {
		layout.form.AddFormItem(formItem.item)
		layout.parameters[formItem.item] = formItem
		if formItem.parameter.ID == config.NetworkID {
			dropDown := formItem.item.(*DropDown)
			dropDown.SetSelectedFunc(func(text string, index int) {
				newNetwork := configPage.home.md.Config.Stadernode.Network.Options[index].Value.(cfgtypes.Network)
				configPage.home.md.Config.ChangeNetwork(newNetwork)
				configPage.home.refresh()
			})
		}
	}
	layout.refresh()

}

// Handle a bulk redraw request
func (configPage *SmartnodeConfigPage) handleLayoutChanged() {
	configPage.layout.refresh()
}
