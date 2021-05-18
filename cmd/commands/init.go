package commands

import (
	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
	"github.com/spf13/cobra"
)

func InitDBCommandContextCobra(command *cobra.Command) (*app.App, error) {
	config, err := command.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	a, err := InitDBCommandContext(config)

	if err != nil {
		// Returning an error just prints the usage message, so actually panic
		panic(err)
	}

	//a.InitPlugins(*a.Config().PluginSettings.Directory, *a.Config().PluginSettings.ClientDirectory)
	//a.DoAdvancedPermissionsMigration()
	//fmt.Println("sssssssssssssssssssssssssssssssssss")
	//a.DoEmojisPermissionsMigration()

	return a, nil
}

func InitDBCommandContext(configDSN string) (*app.App, error) {
	//fmt.Println("model.BuildEnterpriseReady", model.BuildEnterpriseReady)
	if err := utils.TranslationsPreInit(); err != nil {
		return nil, err
	}
	model.AppErrorInit(utils.T)

	s, err := app.NewServer(app.Config(configDSN, false))
	if err != nil {
		return nil, err
	}

	a := s.FakeApp()

	//fmt.Println("model.BuildEnterpriseReady", model.BuildEnterpriseReady)

	if model.BuildEnterpriseReady == "true" {
		a.LoadLicense()
	}

	return a, nil
}