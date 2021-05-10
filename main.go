package main

import (
	"flag"

	"github.com/ghostbaby/cfs-broker/pkg/router"

	"github.com/ghostbaby/cfs-broker/pkg/reload"

	"github.com/ghostbaby/cfs-broker/pkg/g"
	ctrl "sigs.k8s.io/controller-runtime"
)

// @title cfs-broker
// @version 1.0
// @description CFS调试节点
// @termsOfService https://github.com/ghostbaby/cfs-broker
// @in header
// @license.name Herman Zhu

func main() {

	//log := ctrl.Log.WithName("controllers").WithName("cfs-broker")
	cfg := flag.String("config", "", "configuration file")
	flag.Parse()
	g.ParseConfig(*cfg, false)

	log := ctrl.Log.WithName("controllers").WithName("hubble-container-monitor")

	go reload.ConfigReload(log)
	router := router.InitRouter(log)
	router.Run(g.Config().Http.Listen)

	select {}

}
