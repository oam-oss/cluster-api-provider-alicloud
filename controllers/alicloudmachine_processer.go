package controllers

import (
	//"bytes"
	//"compress/gzip"
	rawctx "context"
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/cluster-api-provider-alicloud/pkg"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/juju/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg/aliyun"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	DefaultOSImageId    = "aliyun_2_1903_64_20G_alibase_20190829.vhd"
	DefaultInstanceType = "ecs.c1.large"
)

type MachineProcesser struct {
	AlicloudMachineReconciler
	cluster      *clusterv1.Cluster
	machine      *clusterv1.Machine
	clusterInfra *infrav1.AlicloudCluster
	machineInfra *infrav1.AlicloudMachine
	result       ctrl.Result
	err          error
	ecsEnginer   *ecs.Client
	slbEnginer   *aliyun.SLBClient
	ecsInstance  *ecs.Instance
	isChange     bool
	pather       *patch.Helper
}

func (p *MachineProcesser) init() {
	if p.err != nil {
		return
	}
	ecsClient, err := aliyun.GetEcsClient(p.Info().RegionId())
	if err != nil {
		p.err = err
		return
	}
	p.ecsEnginer = ecsClient

	slbClient, err := aliyun.NewSLBClient(p.Log, p.Info().RegionId())
	if err != nil {
		p.err = err
		return
	}
	p.slbEnginer = slbClient

	if patcher, err := patch.NewHelper(p.machineInfra, p.Client); err == nil {
		p.pather = patcher
	} else {
		p.Log.V(2).Info("patch init...")
	}
}

func (p *MachineProcesser) gobreak(err error) {
	p.err = err
	return
}

func (p *MachineProcesser) goRetry(dur time.Duration) {
	if p.result.Requeue == true {
		if p.result.RequeueAfter.Seconds() > dur.Seconds() {
			p.result.RequeueAfter = dur
		}
	} else {
		p.result.Requeue = true
		p.result.RequeueAfter = dur
	}
	return
}

func (p *MachineProcesser) tryGetInstance() error {
	info := p.Info()
	if len(info.id()) > 0 {
		req := ecs.CreateDescribeInstancesRequest()
		req.RegionId = info.RegionId()
		req.InstanceIds = fmt.Sprintf(`["%s"]`, info.id())
		p.Log.Info("DescribeInstances Request", "request", req)
		response, err := p.ecsEnginer.DescribeInstances(req)
		if err != nil {
			return err
		}
		p.Log.Info("Get Instance", "resp", response)
		if len(response.Instances.Instance) > 0 {
			p.Log.Info("recive EcsInstance: id=" + response.Instances.Instance[0].InstanceId)
			p.ecsInstance = &response.Instances.Instance[0]
		}

	}
	return nil
}

func (p *MachineProcesser) setProviderID(v string) {
	p.machineInfra.Spec.ProviderID = fmt.Sprintf("aliyun://%s", v)
	p.isChange = true
}

type instanceCreateOption func(*ecs.RunInstancesRequest)

func (p *MachineProcesser) createInstance(opts ...instanceCreateOption) error {

	req := ecs.CreateRunInstancesRequest()

	for _, f := range opts {
		f(req)
	}
	p.Log.Info("create instance ", "request", req)

	reponse, err := p.ecsEnginer.RunInstances(req)
	if err != nil {
		p.Log.Error(err, "create ecs instance")
		return err
	}
	if len(reponse.InstanceIdSets.InstanceIdSet) > 0 {
		id := reponse.InstanceIdSets.InstanceIdSet[0]
		p.Log.Info("set id ", "id", id)
		p.Info().setId(id)
	}
	return nil
}

func (p *MachineProcesser) deleteInstance() error {
	req := ecs.CreateDeleteInstanceRequest()
	req.Force = requests.NewBoolean(true)
	req.InstanceId = p.Info().id()
	req.RegionId = p.Info().RegionId()
	if _, err := p.ecsEnginer.DeleteInstance(req); err != nil {
		p.Log.Error(err, "delete instance error", "InstanceId", req.InstanceId)
		return err
	}
	p.Log.Info("delete instance ok", "InstanceId", req.InstanceId)
	return nil
}

func (p *MachineProcesser) bindFinalizers() {
	if !util.Contains(p.machineInfra.Finalizers, infrav1.ClusterFinalizer) {
		p.machineInfra.Finalizers = append(p.machineInfra.Finalizers, infrav1.ClusterFinalizer)
		p.isChange = true
	}
}

func (p *MachineProcesser) unbindFinalizers() {
	if util.Contains(p.machineInfra.Finalizers, infrav1.ClusterFinalizer) {
		p.machineInfra.Finalizers = util.Filter(p.machineInfra.Finalizers, infrav1.ClusterFinalizer)
		p.isChange = true
	}
}

func (p *MachineProcesser) reconcileSLBEndpoint() error {
	lbID := p.clusterInfra.Status.Network.SLB.VServerGroupId
	_, err := p.slbEnginer.VGAddBackendServers(lbID, p.Info().id(), "6443", p.ecsInstance.HostName)
	if err != nil {
		return err
	}
	slbID := p.clusterInfra.Status.Network.SLB.LoadBalancerId
	return p.slbEnginer.StartListener(slbID)
}

func (p *MachineProcesser) commit() {
	if p.isChange {
		p.Log.Info("commit patch ", "id", p.machineInfra.Status.ID)
		//p.Log.Info(fmt.Sprintf("unbind Finalize; now is %v", p.machineInfra.Finalizers))
		if err := p.pather.Patch(rawctx.Background(), p.machineInfra); err != nil {
			p.Log.Error(err, "patch infra-machine")
			//p.deleteInstance()
			p.gobreak(err)
		}
	}
}

func (p *MachineProcesser) handleDelete() {
	p.Log.Info("Handling MachineInfra Delete")
	defer func() {
		if !p.result.Requeue && p.err == nil {
			p.unbindFinalizers()
		}
	}()

	if p.Info().id() != "" {
		if err := p.tryGetInstance(); err != nil {
			p.Log.Error(err, "get ecs instance error when handle delete")
			p.goRetry(time.Second * 20)
			return
		}

		if p.ecsInstance == nil {
			p.Log.Info("ecs instance maybe removed,last try")
			p.deleteInstance()
			return
		}

		switch p.ecsInstance.Status {
		case "Stopping", "Stopped":
			p.Log.Info("ecs instance is stopping|stopped")
		default:
			p.Log.Info("Deletting Ecs Instance")
			if err := p.deleteInstance(); err != nil {
				p.gobreak(errors.Annotate(err, "Delete ECS api error"))
				return
			}

		}
	}

}
func (p *MachineProcesser) sync() {
	if p.err != nil {
		return
	}

	defer p.commit()

	p.Log.Info("AlicloudMachine Sync...")

	if p.machineInfra.DeletionTimestamp != nil {
		p.handleDelete()
		return
	}

	if !p.clusterInfra.Status.Ready {
		p.Log.Info("ClusterInfrastructure status not ready")
		return
	}

	if p.machineInfra.Status.ErrorMessage != "" || p.machineInfra.Status.ErrorReason != "" {
		p.Log.Info("machineInfra status error,skip process")
		return
	}

	if p.machine.Spec.Bootstrap.Data == nil {
		p.Log.Info("machine bootstrap not set")
		return
	}

	p.bindFinalizers()

	info := p.Info()
	//if info.IsMachineReady() {
	//	return
	//}

	if info.id() == "" {

		p.Log.Info("id is null, so create instance")
		if err := p.createInstance(func(req *ecs.RunInstancesRequest) {
			if err := info.FillRunInstancesReq(req); err != nil {
				p.Log.Error(err, "Fill RunInstances Request")
			}
		}); err != nil {
			p.Log.Error(err, "create ecs instance")
			p.goRetry(time.Second * 30)
			return
		}
	}

	if err := p.tryGetInstance(); err != nil {
		p.Log.Error(err, "tryGetInstance")
		p.gobreak(err)
		return
	}

	// ecs get instance after create is null
	if p.ecsInstance == nil {
		p.Log.Info("p.ecsInstance is nil go retry...")
		p.goRetry(time.Second * 5)
		return
	}

	info.updateMachineStatus(func(status *infrav1.AlicloudMachineStatus) {

		status.Addresses = info.getAddresses()
		status.Instance = info.instance()

		p.setProviderID(status.Instance.InstanceId)
		if len(status.Addresses) > 0 && status.Instance.Status == "Running" {
			status.Ready = true

			if info.IsControlPlane() {
				p.Log.Info("machine is controlplane, reconcileSLBEndpoint...")
				if err := p.reconcileSLBEndpoint(); err != nil {
					p.Log.Error(err, "reconcileSLBEndpoint")
					p.goRetry(time.Second * 10)
					status.Ready = false
				}
			}
		}

		if !status.Ready {
			p.goRetry(time.Second * 15)
		}
	})

}

func (p *MachineProcesser) Info() *InfoProvider {
	return &InfoProvider{
		store: p,
	}
}

type InfoProvider struct {
	store *MachineProcesser
}

func (s *InfoProvider) RegionId() string {
	return s.store.clusterInfra.Spec.RegionId
}

func (s *InfoProvider) ZoneId() string {
	return s.store.clusterInfra.Spec.ZoneId
}

func (s *InfoProvider) VSwitchId() string {
	return s.store.clusterInfra.Status.Network.VSwitch.VSwitchId
}

func (s *InfoProvider) MachineName() string {
	return s.store.machine.Name
}

func (s *InfoProvider) SecurityGroupId() string {
	return s.store.clusterInfra.Status.Network.SecurityGroup.SecurityGroupId
}

func (s *InfoProvider) IsMachineReady() bool {
	return s.store.machineInfra.Status.Ready
}

func (s *InfoProvider) id() string {
	infra := s.store.machineInfra
	if infra.Status.ID != "" {
		return infra.Status.ID
	}
	key := client.ObjectKey{Name: infra.Name, Namespace: infra.Namespace}
	err := s.store.Client.Get(rawctx.TODO(), key, s.store.machineInfra)
	if err != nil {
		return ""
	}
	return s.store.machineInfra.Status.ID
}

func (s *InfoProvider) getProviderID() string {
	rid := s.store.machineInfra.Spec.ProviderID
	if !strings.Contains(rid, "://") {
		return ""
	}
	lastSlashIndex := strings.LastIndex(rid, "/")
	return rid[lastSlashIndex+1:]
}

func (s *InfoProvider) updateMachineStatus(f func(status *infrav1.AlicloudMachineStatus)) {
	f(&s.store.machineInfra.Status)
	s.store.isChange = true
}

func (s *InfoProvider) getAddresses() (machineAddresses clusterv1.MachineAddresses) {
	for _, inface := range s.store.ecsInstance.NetworkInterfaces.NetworkInterface {
		machineAddresses = append(machineAddresses, clusterv1.MachineAddress{
			Type:    clusterv1.MachineInternalIP,
			Address: inface.PrimaryIpAddress,
		})
	}

	for _, address := range s.store.ecsInstance.PublicIpAddress.IpAddress {
		machineAddresses = append(machineAddresses, clusterv1.MachineAddress{
			Type:    clusterv1.MachineExternalIP,
			Address: address,
		})
	}

	return
}

func (s *InfoProvider) instance() *infrav1.Instance {
	return infrav1.InstanceFromEcs(s.store.ecsInstance)
}

func (s *InfoProvider) setId(v string) {
	s.store.machineInfra.Status.ID = v
	s.store.isChange = true
}

func (s *InfoProvider) IsControlPlane() bool {
	return util.IsControlPlaneMachine(s.store.machine)
}

func (s *InfoProvider) FillRunInstancesReq(req *ecs.RunInstancesRequest) error {
	req.RegionId = s.RegionId()
	req.ZoneId = s.ZoneId()
	req.VSwitchId = s.VSwitchId()
	req.InstanceName = s.MachineName()
	req.SecurityGroupId = s.SecurityGroupId()
	req.MinAmount = requests.NewInteger(1)
	req.Amount = requests.NewInteger(1)
	//if s.IsControlPlane() {
	//req.InternetMaxBandwidthOut = requests.NewInteger(1)
	//req.InternetMaxBandwidthIn = requests.NewInteger(1)
	//}
	req.KeyPairName = "codedeploy"
	fillInstanceReqByMachineSpec(req, s.store.machineInfra.Spec)

	bootstrapData := s.store.machine.Spec.Bootstrap.Data
	if bootstrapData != nil {
		req.UserData = *bootstrapData
		//var buf bytes.Buffer
		//
		//decoded, err := base64.StdEncoding.DecodeString(*bootstrapData)
		//if err != nil {
		//	return errors.Annotate(err, "failed to decode bootstrapData")
		//}
		//
		//gz := gzip.NewWriter(&buf)
		//if _, err := gz.Write(decoded); err != nil {
		//	return errors.Annotate(err, "failed to gzip userdata")
		//}
		//
		//if err := gz.Close(); err != nil {
		//	return errors.Annotate(err, "failed to gzip userdata")
		//}
		//
		//req.UserData = base64.StdEncoding.EncodeToString(buf.Bytes())
	}
	return nil
}

func fillInstanceReqByMachineSpec(req *ecs.RunInstancesRequest, spec infrav1.AlicloudMachineSpec) {
	req.InternetChargeType = spec.InternetChargeType
	if len(spec.CapacityReservationId) > 0 {
		req.CapacityReservationId = spec.CapacityReservationId
	}
	req.KeyPairName = spec.SSHKeyPair
	if req.KeyPairName == "" {
		req.KeyPairName = pkg.DefaultSSHKeyName
	}
	req.ImageId = spec.ImageId
	if len(req.ImageId) == 0 {
		req.ImageId = DefaultOSImageId
	}
	req.InstanceType = spec.InstanceType
	if len(req.InstanceType) == 0 {
		req.InstanceType = DefaultInstanceType
	}

	if len(spec.InternetMaxBandwidthIn) > 0 && len(spec.InternetMaxBandwidthOut) > 0 {
		req.InternetMaxBandwidthIn = requests.Integer(spec.InternetMaxBandwidthIn)
		req.InternetMaxBandwidthOut = requests.Integer(spec.InternetMaxBandwidthOut)
	}

	if len(spec.SystemDiskCategory) > 0 {
		req.SystemDiskCategory = spec.SystemDiskCategory
	}
	req.SystemDiskSize = spec.SystemDiskSize
}
