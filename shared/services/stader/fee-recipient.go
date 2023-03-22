package stader

// ROCKETPOOL-OWNED

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stader-labs/stader-node/shared/services/config"
)

// Config
const (
	FileMode fs.FileMode = 0644
)

// Checks if the fee recipient file exists and has the correct distributor address in it.
// The first return value is for file existence, the second is for validation of the fee recipient address inside.
func CheckFeeRecipientFile(feeRecipient common.Address, cfg *config.StaderConfig) (bool, bool, error) {

	// Check if the file exists
	path := cfg.StaderNode.GetFeeRecipientFilePath()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	// Compare the file contents with the expected string
	expectedString := getFeeRecipientFileContents(feeRecipient, cfg)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return false, false, fmt.Errorf("error reading fee recipient file: %w", err)
	}
	existingString := string(bytes)
	if existingString != expectedString {
		// If it wrote properly, indicate a success but that the file needed to be updated
		return true, false, nil
	}

	// The file existed and had the expected address, all set.
	return true, true, nil
}

// Writes the given address to the fee recipient file. The VC should be restarted to pick up the new file.
func UpdateFeeRecipientFile(feeRecipient common.Address, cfg *config.StaderConfig) error {

	// Create the distributor address string for the node
	expectedString := getFeeRecipientFileContents(feeRecipient, cfg)
	bytes := []byte(expectedString)

	// Write the file
	path := cfg.StaderNode.GetFeeRecipientFilePath()
	err := ioutil.WriteFile(path, bytes, FileMode)
	if err != nil {
		return fmt.Errorf("error writing fee recipient file: %w", err)
	}
	return nil

}

// Gets the expected contents of the fee recipient file
func getFeeRecipientFileContents(feeRecipient common.Address, cfg *config.StaderConfig) string {
	if !cfg.IsNativeMode {
		// Docker mode
		return feeRecipient.Hex()
	}

	// Native mode
	return fmt.Sprintf("%s=%s", config.FeeRecipientEnvVar, feeRecipient.Hex())
}
