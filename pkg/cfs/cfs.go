package cfs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	traitv1 "github.com/ghostbaby/cfs-broker/pkg/api/v1"
	"github.com/ghostbaby/cfs-broker/pkg/libdocker"
	"k8s.io/klog"
)

const (
	DefaultCgroupPath          = "/data/kubepods"
	DefaultCfsPeriodUsFileName = "cpu.cfs_period_us"
	DefaultCfsQuotaUsFileName  = "cpu.cfs_quota_us"
)

func Validate(cfs *traitv1.CfsTrait) error {
	if (!cfs.Spec.IsAllPods && len(cfs.Spec.Pods) == 0) ||
		(cfs.Spec.IsAllPods && len(cfs.Spec.Pods) > 0) {
		return errors.New("spec.isAllPods and spec.pods can only configure one of the parameters")
	}
	return nil
}

func Controller(cfs *traitv1.CfsTrait) error {

	cli := libdocker.ConnectToDockerOrDie(0)
	opts := dockertypes.ContainerListOptions{}
	containers, err := cli.ListContainers(opts)
	if err != nil {
		klog.Error("fail to get containers, err:", err)
		return err
	}
	for _, container := range containers {
		var (
			cPath, podPath string
		)

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
			podPath = strings.Join([]string{DefaultCgroupPath, cParent}, "/")
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

		if cfsPeriod == cfs.Spec.Period && cfsQuota == cfs.Spec.Quota {
			klog.Infof("app %s, cfs value not change, skip to sync", cfs.Spec.AppName)
			continue
		}

		if !cfs.Spec.Force {

			expect := float64(cfs.Spec.Quota) / float64(cfs.Spec.Period)
			current := float64(cfsQuota) / float64(cfsPeriod)

			if expect < current {
				klog.Errorf("app %s, expect value %f < current value %f, skip to exec",
					cfs.Spec.AppName, expect, current)
				continue
			}
		}

		if cfsQuota <= cfs.Spec.Quota {

			if err := WriteContainerCfsConfig(podPath, DefaultCfsQuotaUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Quota)); err != nil {

				klog.Errorf("fail to write pod %s cfs quota us", metadata.PodName)
				continue
			}

			if err := WriteContainerCfsConfig(cPath, DefaultCfsQuotaUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Quota)); err != nil {

				klog.Errorf("fail to write container %s cfs quota us", metadata.PodName)
				continue
			}
		}

		if cfsQuota > cfs.Spec.Quota {
			if err := WriteContainerCfsConfig(cPath, DefaultCfsQuotaUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Quota)); err != nil {

				klog.Errorf("fail to write container %s cfs quota us", metadata.PodName)
				continue
			}

			if err := WriteContainerCfsConfig(podPath, DefaultCfsQuotaUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Quota)); err != nil {

				klog.Errorf("fail to write pod %s cfs quota us", metadata.PodName)
				continue
			}
		}

		if cfsPeriod <= cfs.Spec.Period {

			if err := WriteContainerCfsConfig(cPath, DefaultCfsPeriodUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Period)); err != nil {

				klog.Errorf("fail to write container %s cfs period us", metadata.PodName)
				continue
			}

			if err := WriteContainerCfsConfig(podPath, DefaultCfsPeriodUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Period)); err != nil {

				klog.Errorf("fail to write pod %s cfs period us", metadata.PodName)
				continue
			}
		}

		if cfsPeriod > cfs.Spec.Period {
			if err := WriteContainerCfsConfig(podPath, DefaultCfsPeriodUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Period)); err != nil {

				klog.Errorf("fail to write pod %s cfs period us", metadata.PodName)
				continue
			}

			if err := WriteContainerCfsConfig(cPath, DefaultCfsPeriodUsFileName, metadata.PodName,
				fmt.Sprintf("%d", cfs.Spec.Period)); err != nil {

				klog.Errorf("fail to write container %s cfs period us", metadata.PodName)
				continue
			}
		}

		klog.Infof("success to change cfs arguments, app: %s , cfs_period_us: %d, cfs_quota_us: %d",
			cfs.Spec.AppName, cfs.Spec.Period, cfs.Spec.Quota)
	}
	return nil
}

func GetContainerCfsConfig(path, cfsFile, podName string) (int32, error) {
	cfsFilePath := strings.Join([]string{path, cfsFile}, "/")
	value, err := ReadCfsConfig(cfsFilePath)
	if err != nil {
		klog.Errorf("fail to get container %s cfs value from %s, err: %v", podName, cfsFile, err)
		return 0, err
	}
	return value, nil
}

func WriteContainerCfsConfig(path, cfsFile, podName, value string) error {
	cfsFilePath := strings.Join([]string{path, cfsFile}, "/")
	if err := WriteCfsConfig(cfsFilePath, value); err != nil {
		klog.Errorf("fail to write container %s cfs value to %s, err: %v", podName, cfsFile, err)
		return err
	}
	return nil
}

func ReadCfsConfig(path string) (int32, error) {

	f, err := os.Open(path)
	if err != nil {
		klog.Errorf("read file fail %s,err:%v", path, err)
		return 0, err
	}
	defer f.Close()

	fd, err := ioutil.ReadAll(f)
	if err != nil {
		klog.Errorf("read to fd fail %s,err:%v", path, err)
		return 0, err
	}

	parseInt, err := strconv.ParseInt(strings.Replace(string(fd), "\n", "", -1), 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(parseInt), nil
}

func WriteCfsConfig(path, value string) error {

	err := ioutil.WriteFile(path, []byte(value), 0755)
	if err != nil {
		klog.Error("ioutil WriteFile error: ", err)
		return err
	}
	return nil
}
