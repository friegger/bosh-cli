package cmd

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/validation"
)

type deployCmd struct {
	ui        bmui.UI
	config    bmconfig.Config
	fs        boshsys.FileSystem
	extractor bmtar.Extractor
}

func NewDeployCmd(
	ui bmui.UI,
	config bmconfig.Config,
	fs boshsys.FileSystem,
	extractor bmtar.Extractor,
) *deployCmd {
	return &deployCmd{
		ui:        ui,
		config:    config,
		fs:        fs,
		extractor: extractor,
	}
}

func (c *deployCmd) Run(args []string) error {
	if len(args) == 0 {
		c.ui.Error("No CPI release provided")
		return errors.New("No CPI release provided")
	}

	cpiPath := args[0]
	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(cpiPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release '%s' does not exist", cpiPath))
		return bosherr.WrapError(err, "Checking CPI release '%s' existence", cpiPath)
	}

	if len(c.config.Deployment) == 0 {
		c.ui.Error("No deployment set")
		return errors.New("No deployment set")
	}

	err = fileValidator.Exists(c.config.Deployment)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment manifest path '%s' does not exist", c.config.Deployment))
		return bosherr.WrapError(err, "Reading deployment manifest for deploy")
	}

	extractedReleasePath, err := c.fs.TempDir("cmd-deployCmd")
	if err != nil {
		c.ui.Error("Could not create a temporary directory")
		return bosherr.WrapError(err, "Creating extracted release path")
	}
	defer c.fs.RemoveAll(extractedReleasePath)

	releaseReader := bmrelease.NewTarReader(cpiPath, extractedReleasePath, c.fs, c.extractor)
	release, err := releaseReader.Read()
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release '%s' is not a BOSH release", cpiPath))
		return bosherr.WrapError(err, fmt.Sprintf("Reading CPI release from '%s'", cpiPath))
	}

	validator := bmrelease.NewValidator(c.fs, release)
	err = validator.Validate()
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release '%s' is not a valid BOSH release", cpiPath))
		return bosherr.WrapError(err, "Validating CPI release")
	}

	return nil
}
