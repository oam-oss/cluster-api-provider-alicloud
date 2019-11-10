package aliyun

import (
	"os"

	ecssvc "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/denverdino/aliyungo/cs"
	"github.com/denverdino/aliyungo/ecs"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

var (
	AccessKeyId     = os.Getenv("ACCESS_KEY_ID")
	AccessKeySecret = os.Getenv("ACCESS_SECRET")
)

func GetEcsClient(regionId string) (*ecssvc.Client, error) {
	return ecssvc.NewClientWithAccessKey(regionId, AccessKeyId, AccessKeySecret)
}

func CreationArgsBySpec(cluster clusterv1.Cluster) *cs.KubernetesCreationArgs {
	spec := cluster.Spec

	creationArgs := &cs.KubernetesCreationArgs{
		Name:                     cluster.Name,
		ClusterType:              "Kubernetes",
		DisableRollback:          true,
		TimeoutMins:              60,
		MasterInstanceType:       "",
		WorkerInstanceType:       "",
		VPCID:                    "",
		VSwitchId:                "",
		LoginPassword:            "",
		KeyPair:                  "",
		Network:                  "vrouter",
		NodeCIDRMask:             "",
		LoggingType:              "",
		SLSProjectName:           "",
		NumOfNodes:               0,
		MasterSystemDiskCategory: ecs.DiskCategory(""),
		MasterSystemDiskSize:     int64(1 << 9),
		WorkerSystemDiskCategory: ecs.DiskCategory(""),
		WorkerSystemDiskSize:     int64(1 << 9),
		SNatEntry:                true,
		KubernetesVersion:        "1.15.4",
		ContainerCIDR:            spec.ClusterNetwork.Pods.CIDRBlocks[0],
		ServiceCIDR:              spec.ClusterNetwork.Services.CIDRBlocks[0],
		SSHFlags:                 true,
		PublicSLB:                true,
		CloudMonitorFlags:        true,
		ZoneId:                   "",
	}

	return creationArgs
}

func ProcessByMd(args *cs.KubernetesCreationArgs, deployment clusterv1.MachineDeployment) {

}
