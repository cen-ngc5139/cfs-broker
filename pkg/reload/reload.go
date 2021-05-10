package reload

import (
	"time"

	"github.com/ghostbaby/cfs-broker/pkg/g"
	"github.com/go-logr/logr"
)

func reloadConfig(log logr.Logger) {
	g.ParseConfig(g.ConfigFile, true)
	log.Info("reload config complete")

}

func ConfigReload(log logr.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		reloadConfig(log)
		<-ticker.C
	}
}
