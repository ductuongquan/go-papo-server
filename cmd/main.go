// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"bitbucket.org/enesyteam/papo-server/cmd/commands"
	// Plugins
	_ "bitbucket.org/enesyteam/papo-server/model/facebook"
	_ "github.com/go-ldap/ldap"
	_ "github.com/hako/durafmt"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/tylerb/graceful"
	"os"

)

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
