// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.s

package commands

import (
	"bitbucket.org/enesyteam/papo-server/api1"
	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/web"
	"bitbucket.org/enesyteam/papo-server/wsapi"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	SESSIONS_CLEANUP_BATCH_SIZE = 1000
)

var MaxNotificationsPerChannelDefault int64 = 1000000

var serverCmd = &cobra.Command{
	Use:          "server",
	Short:        "Run the Papo server",
	RunE:         serverCmdF,
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(serverCmd)
	RootCmd.RunE = serverCmdF
}

func serverCmdF(command *cobra.Command, args []string) error {
	config, err := command.Flags().GetString("config")
	if err != nil {
		return err
	}

	disableConfigWatch, _ := command.Flags().GetBool("disableconfigwatch")
	usedPlatform, _ := command.Flags().GetBool("platform")

	interruptChan := make(chan os.Signal, 1)
	return runServer(config, disableConfigWatch, usedPlatform, interruptChan)
}

func runServer(configDSN string, disableConfigWatch bool, usedPlatform bool, interruptChan chan os.Signal) error {
	options := []app.Option{
		app.Config(configDSN, !disableConfigWatch),
		app.RunJobs,
		app.JoinCluster,
		app.StartElasticsearch,
		app.StartMetrics,
	}
	server, err := app.NewServer(options...)
	if err != nil {
		mlog.Critical(err.Error())
		return err
	}
	defer server.Shutdown()

	if usedPlatform {
		mlog.Error("The platform binary has been deprecated, please switch to using the mattermost binary.")
	}

	serverErr := server.Start()
	if serverErr != nil {
		mlog.Critical(serverErr.Error())
		return serverErr
	}

	api1.Init(server, server.AppOptions, server.Router)
	wsapi.Init(server.FakeApp(), server.WebSocketRouter)
	web.New(server, server.AppOptions, server.Router)

	// If we allow testing then listen for manual testing URL hits
	//if *server.Config().ServiceSettings.EnableTesting {
	//	manualtesting.Init(api)
	//}

	notifyReady()

	// wait for kill signal before attempting to gracefully shutdown
	// the running service
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

	return nil
}

func runSecurityJob(a *app.App) {
	doSecurity(a)
	model.CreateRecurringTask("Security", func() {
		doSecurity(a)
	}, time.Hour*4)
}

func runDiagnosticsJob(a *app.App) {
	doDiagnostics(a)
	model.CreateRecurringTask("Diagnostics", func() {
		doDiagnostics(a)
	}, time.Hour*24)
}

func runTokenCleanupJob(a *app.App) {
	doTokenCleanup(a)
	model.CreateRecurringTask("Token Cleanup", func() {
		doTokenCleanup(a)
	}, time.Hour*1)
}

func runCommandWebhookCleanupJob(a *app.App) {
	doCommandWebhookCleanup(a)
	model.CreateRecurringTask("Command Hook Cleanup", func() {
		doCommandWebhookCleanup(a)
	}, time.Hour*1)
}

func runSessionCleanupJob(a *app.App) {
	doSessionCleanup(a)
	model.CreateRecurringTask("Session Cleanup", func() {
		doSessionCleanup(a)
	}, time.Hour*24)
}

func resetStatuses(a *app.App) {
	//if result := <-a.Srv.Store.Status().ResetAll(); result.Err != nil {
	//	mlog.Error(fmt.Sprint("Error to reset the server status.", result.Err.Error()))
	//}
}

func doSecurity(a *app.App) {
	//a.DoSecurityUpdateCheck()
}

func doDiagnostics(a *app.App) {
	if *a.Config().LogSettings.EnableDiagnostics {
		//a.SendDailyDiagnostics()
	}
}

func notifyReady() {
	// If the environment vars provide a systemd notification socket,
	// notify systemd that the server is ready.
	systemdSocket := os.Getenv("NOTIFY_SOCKET")
	if systemdSocket != "" {
		mlog.Info("Sending systemd READY notification.")

		err := sendSystemdReadyNotification(systemdSocket)
		if err != nil {
			mlog.Error(err.Error())
		}
	}
}

func sendSystemdReadyNotification(socketPath string) error {
	msg := "READY=1"
	addr := &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg))
	return err
}

func doTokenCleanup(a *app.App) {
	a.Srv.Store.Token().Cleanup()
}

func doCommandWebhookCleanup(a *app.App) {
	//a.Srv.Store.CommandWebhook().Cleanup()
}

func doSessionCleanup(a *app.App) {
	a.Srv.Store.Session().Cleanup(model.GetMillis(), SESSIONS_CLEANUP_BATCH_SIZE)
}
