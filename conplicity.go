package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/engines"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/providers"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
)

var version = "undefined"

func main() {
	var err error
	var exitCode int = 0

	c, err := handler.NewConplicity(version)
	util.CheckErr(err, "Failed to setup Conplicity handler: %v", "fatal")

	log.Infof("Conplity v%s starting backup...", version)

	vols, err := c.GetVolumes()
	util.CheckErr(err, "Failed to get Docker volumes: %v", "fatal")

	for _, vol := range vols {
		metrics, err := backupVolume(c, vol)
		if err != nil {
			log.Errorf("Failed to backup volume %s: %v", vol.Name, err)
			exitCode = 1
			continue
		}
		c.Metrics = append(c.Metrics, metrics...)
	}

	err = c.PushToPrometheus()
	if err != nil {
		log.Errorf("Failed to post data to Prometheus Pushgateway: %v", err)
		exitCode = 2
	}

	log.Infof("End backup...")
	os.Exit(exitCode)
}

func backupVolume(c *handler.Conplicity, vol *volume.Volume) (metrics []string, err error) {
	p := providers.GetProvider(c, vol)
	log.WithFields(log.Fields{
		"volume":   vol.Name,
		"provider": p.GetName(),
	}).Info("Found data provider")
	err = providers.PrepareBackup(p)
	util.CheckErr(err, "Failed to prepare backup for volume "+vol.Name+": %v", "fatal")

	e := engines.GetEngine(c, vol)
	log.WithFields(log.Fields{
		"volume": vol.Name,
		"engine": e.GetName(),
	}).Info("Found backup engine")

	metrics, err = e.Backup()
	util.CheckErr(err, "Failed to backup volume "+vol.Name+": %v", "fatal")
	return
}
