package cfs

import (
	"fmt"
	"strings"

	traitv1 "github.com/ghostbaby/cfs-broker/pkg/api/v1"
	"github.com/ghostbaby/cfs-broker/pkg/libdocker"
	"github.com/ghostbaby/cfs-broker/pkg/models"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"k8s.io/klog"
)

func Query(log logr.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		spec := &traitv1.CfsTrait{}
		if err := c.Bind(spec); HandleError(c, err) {
			log.Error(err, "fail to bind post param")
			return
		}

		if err := Validate(spec); HandleError(c, err) {
			log.Error(err, "fail to check post param")
			return
		}

		cfs, err := GetCurrentCfsConfig(spec)
		if HandleError(c, err) {
			log.Error(err, "fail to update cfs config")
			return
		}

		JsonResult(c, cfs)
	}
}

func GetCurrentCfsConfig(cfs *traitv1.CfsTrait) ([]*models.CurrentCfsConfig, error) {
	cli := libdocker.ConnectToDockerOrDie(0)
	opts := dockertypes.ContainerListOptions{}
	containers, err := cli.ListContainers(opts)
	if err != nil {
		klog.Error("fail to get containers, err:", err)
		return nil, err
	}
	var cfsList []*models.CurrentCfsConfig
	for _, container := range containers {
		var cPath string

		inspect, err := cli.InspectContainer(container.ID)
		if err != nil {
			klog.Errorf("fail to get container inspec, err:", err)
			continue
		}

		metadata, err := libdocker.ParseContainerName(container.Names[0])
		if err != nil {
			klog.Errorf("fail to parse container name, err:", err)
			continue
		}

		if metadata.Name == "POD" {
			klog.Infof("skip to parse pause container, id: %s", container.Names[0])
			continue
		}

		if metadata.Name != cfs.Spec.AppName {
			continue
		}

		if metadata.Namespace != cfs.Spec.Namespace {
			continue
		}

		if (!cfs.Spec.IsAllPods && len(cfs.Spec.Pods) == 0) ||
			(cfs.Spec.IsAllPods && len(cfs.Spec.Pods) > 0) {
			klog.Infof("spec.isAllPods and spec.pods can only configure one of the parameters")
			continue
		}

		if len(cfs.Spec.Pods) > 0 {
			var ok bool
			for _, pod := range cfs.Spec.Pods {
				if metadata.PodName == pod {
					ok = true
				}
			}

			if !ok {
				klog.Infof("cannot find the container in spec.pods on the current node")
				continue
			}
		}

		cParent := inspect.HostConfig.CgroupParent

		if len(cParent) == 0 {
			klog.Infof("container %s CgroupParent not exist,skip....", metadata.Name)
			continue
		}

		if strings.Contains(cParent, "kubepods-burstable") ||
			strings.Contains(cParent, "kubepods-besteffort") {
			//path = strings.Join([]string{DefaultCgroupPath, DefaultBurstablePath, cParent}, "/")
			klog.Infof("not support burstable or besteffort model")
			continue
		}

		dockerPath := fmt.Sprintf("docker-%s.scope", container.ID)

		if strings.Contains(cParent, "kubepods-pod") {
			cPath = strings.Join([]string{DefaultCgroupPath, cParent, dockerPath}, "/")
		}

		if len(cPath) == 0 {
			continue
		}

		cfsPeriod, err := GetContainerCfsConfig(cPath, DefaultCfsPeriodUsFileName, metadata.PodName)
		if err != nil {
			klog.Errorf("fail to get container %s cfs period us", metadata.PodName)
			continue
		}

		cfsQuota, err := GetContainerCfsConfig(cPath, DefaultCfsQuotaUsFileName, metadata.PodName)
		if err != nil {
			klog.Errorf("fail to get container %s cfs quota us", metadata.PodName)
			continue
		}

		klog.Infof("container %s current cfs arguments, cfs_period_us: %d, cfs_quota_us: %d",
			metadata.PodName, cfsPeriod, cfsQuota)

		cfsList = append(
			cfsList,
			&models.CurrentCfsConfig{
				CfsPeriodUS: cfsPeriod,
				CfsQuotaUS:  cfsQuota,
			},
		)
	}

	return cfsList, nil
}
