package controllers

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg/aliyun"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg/aliyun/retry"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ClusterProcessor struct {
	logr.Logger

	client          client.Client
	patchHelper     *patch.Helper
	cluster         *clusterv1.Cluster
	alicloudCluster *infrav1.AlicloudCluster

	slb           *aliyun.SLBClient
	vpc           *aliyun.VPCClient
	vswitch       *aliyun.VSwitchClient
	securityGroup *aliyun.SecurityGroupClient
}

func NewClusterProcessor(
	logger logr.Logger,
	regionID string,
	client client.Client,
	cluster *clusterv1.Cluster,
	alicloudCluster *infrav1.AlicloudCluster,
) (*ClusterProcessor, error) {
	helper, err := patch.NewHelper(alicloudCluster, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	slbCli, err := aliyun.NewSLBClient(logger, regionID)
	if err != nil {
		return nil, errors.Wrap(err, "NewSLBClient")
	}
	vpcCli, err := aliyun.NewVPCClient(logger, regionID)
	if err != nil {
		return nil, errors.Wrap(err, "NewVPCClient")
	}
	vswitchCli, err := aliyun.NewVSwitchClient(logger, regionID)
	if err != nil {
		return nil, errors.Wrap(err, "NewVSwitchClient")
	}
	securityGroupCli, err := aliyun.NewSecurityGroupClient(logger, regionID)
	if err != nil {
		return nil, errors.Wrap(err, "NewSecurityGroupClient")
	}

	return &ClusterProcessor{
		Logger: logger,

		client:          client,
		patchHelper:     helper,
		cluster:         cluster,
		alicloudCluster: alicloudCluster,

		slb:           slbCli,
		vpc:           vpcCli,
		vswitch:       vswitchCli,
		securityGroup: securityGroupCli,
	}, nil
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterProcessor) Close() error {
	s.Info("Close")
	return errors.Wrap(s.patch(), "patch")
}

func (s *ClusterProcessor) patch() error {
	return errors.Wrap(s.patchHelper.Patch(context.TODO(), s.alicloudCluster), "patchHelper.Patch")
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *ClusterProcessor) ReconcileDelete() (reconcile.Result, error) {
	s.Info("ReconcileDelete")

	if rs, err := s.deleteNetwork(); err != nil {
		return rs, errors.Wrap(err, "deleteNetwork")
	}

	s.alicloudCluster.Finalizers = filter(s.alicloudCluster.Finalizers, infrav1.ClusterFinalizer)

	s.Info("ReconcileDelete success")
	return reconcile.Result{}, nil
}

func filter(list []string, strToFilter string) (newList []string) {
	for _, item := range list {
		if item != strToFilter {
			newList = append(newList, item)
		}
	}
	return
}

func (s *ClusterProcessor) deleteNetwork() (reconcile.Result, error) {
	s.Info("deleteNetwork")

	if rs, err := s.deleteSecurityGroup(); err != nil {
		return rs, errors.Wrap(err, "deleteSecurityGroup")
	}
	if rs, err := s.deleteSLB(); err != nil {
		return rs, errors.Wrap(err, "deleteSLB")
	}
	if rs, err := s.deleteNat(); err != nil {
		return rs, errors.Wrap(err, "deleteNat")
	}
	if rs, err := s.deleteVSwitch(); err != nil {
		return rs, errors.Wrap(err, "deleteVSwitch")
	}
	if rs, err := s.deleteVPC(); err != nil {
		return rs, errors.Wrap(err, "deleteVPC")
	}

	s.Info("deleteNetwork success")
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) deleteSecurityGroup() (reconcile.Result, error) {
	s.Info("deleteSecurityGroup")

	id := s.alicloudCluster.Status.Network.SecurityGroup.SecurityGroupId
	if len(id) == 0 {
		return reconcile.Result{}, nil
	}

	target, err := s.securityGroup.Describe(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Describe")
	}
	if target == nil {
		return reconcile.Result{}, nil
	}

	err = s.securityGroup.Delete(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "deleteSecurityGroup")
	}

	return reconcile.Result{}, retry.Try(retry.DefaultBackOf, func() error {
		target, err := s.securityGroup.Describe(id)
		if err != nil {
			return err
		}
		if target != nil {
			return retry.ErrRetry
		}
		return nil
	})
}

func (s *ClusterProcessor) deleteSLB() (reconcile.Result, error) {
	s.Info("deleteSLB")

	id := s.alicloudCluster.Status.Network.SLB.LoadBalancerId
	if len(id) == 0 {
		return reconcile.Result{}, nil
	}

	target, err := s.slb.Describe(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Describe")
	}
	if target == nil {
		return reconcile.Result{}, nil
	}

	err = s.slb.Delete(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "deleteSLB")
	}

	return reconcile.Result{}, retry.Try(retry.DefaultBackOf, func() error {
		target, err := s.slb.Describe(id)
		if err != nil {
			return err
		}
		if target != nil {
			return retry.ErrRetry
		}
		return nil
	})
}

func (s *ClusterProcessor) deleteNat() (reconcile.Result, error) {
	s.Info("deleteNat")
	eipID := s.alicloudCluster.Status.Network.Nat.EIP.AllocationId
	ngwID := s.alicloudCluster.Status.Network.Nat.NatGateway.NatGatewayId
	if len(eipID) == 0 && len(ngwID) == 0 {
		return reconcile.Result{}, nil
	}

	if len(s.alicloudCluster.Status.Network.Nat.SnatEntryId) > 0 {
		s.Info("DeleteSnatEntry")
		if err := s.vpc.DeleteSnatEntry(&s.alicloudCluster.Status.Network.Nat); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "DeleteSnatEntry")
		}
	}

	if rs, err := s.deleteEIP(); err != nil {
		return rs, errors.Wrap(err, "deleteEIP")
	}
	if rs, err := s.deleteNatGateway(); err != nil {
		return rs, errors.Wrap(err, "deleteNatGateway")
	}
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) deleteEIP() (reconcile.Result, error) {
	s.Info("deleteEIP")

	id := s.alicloudCluster.Status.Network.Nat.EIP.AllocationId
	if len(id) == 0 {
		return reconcile.Result{}, nil
	}

	target, err := s.vpc.DescribeEIP(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Describe")
	}
	if target == nil {
		return reconcile.Result{}, nil
	}

	if target.Status == infrav1.EIPInUse {
		if err := s.vpc.UnassociateEipToNatGateway(&s.alicloudCluster.Status.Network.Nat.EIP, &s.alicloudCluster.Status.Network.Nat.NatGateway); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "UnassociateEipToNatGateway")
		}
		if _, err = s.vpc.WaitEIPStatus(target.AllocationId, infrav1.Available); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "WaitEIPStatus Available")
		}
	}

	err = s.vpc.DeleteEIP(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "deleteVSwitch")
	}

	return reconcile.Result{}, retry.Try(retry.DefaultBackOf, func() error {
		target, err := s.vpc.DescribeEIP(id)
		if err != nil {
			return err
		}
		if target != nil {
			return retry.ErrRetry
		}
		return nil
	})
}

func (s *ClusterProcessor) deleteNatGateway() (reconcile.Result, error) {
	s.Info("deleteNatGateway")

	id := s.alicloudCluster.Status.Network.Nat.NatGateway.NatGatewayId
	if len(id) == 0 {
		return reconcile.Result{}, nil
	}

	target, err := s.vpc.DescribeNatGateway(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Describe")
	}
	if target == nil {
		return reconcile.Result{}, nil
	}

	err = s.vpc.DeleteGateway(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "deleteVSwitch")
	}

	return reconcile.Result{}, retry.Try(retry.DefaultBackOf, func() error {
		target, err := s.vpc.DescribeNatGateway(id)
		if err != nil {
			return err
		}
		if target != nil {
			return retry.ErrRetry
		}
		return nil
	})
}

func (s *ClusterProcessor) deleteVSwitch() (reconcile.Result, error) {
	s.Info("deleteVSwitch")

	id := s.alicloudCluster.Status.Network.VSwitch.VSwitchId
	if len(id) == 0 {
		return reconcile.Result{}, nil
	}

	target, err := s.vswitch.Describe(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Describe")
	}
	if target == nil {
		return reconcile.Result{}, nil
	}

	err = s.vswitch.Delete(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "deleteVSwitch")
	}

	return reconcile.Result{}, retry.Try(retry.DefaultBackOf, func() error {
		target, err := s.vswitch.Describe(id)
		if err != nil {
			return err
		}
		if target != nil {
			return retry.ErrRetry
		}
		return nil
	})
}

func (s *ClusterProcessor) deleteVPC() (reconcile.Result, error) {
	s.Info("deleteVPC")

	id := s.alicloudCluster.Status.Network.VPC.VpcId
	if len(id) == 0 {
		return reconcile.Result{}, nil
	}

	target, err := s.vpc.Describe(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Describe")
	}
	if target == nil {
		return reconcile.Result{}, nil
	}

	err = s.vpc.Delete(id)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "deleteVPC")
	}

	return reconcile.Result{}, retry.Try(retry.DefaultBackOf, func() error {
		target, err := s.vpc.Describe(id)
		if err != nil {
			return err
		}
		if target != nil {
			return retry.ErrRetry
		}
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *ClusterProcessor) ReconcileNormal() (reconcile.Result, error) {
	if s.alicloudCluster.Status.Ready {
		return reconcile.Result{}, nil
	}

	s.Info("ReconcileNormal")

	if !util.Contains(s.alicloudCluster.Finalizers, infrav1.ClusterFinalizer) {
		s.alicloudCluster.Finalizers = append(s.alicloudCluster.Finalizers, infrav1.ClusterFinalizer)
	}

	if len(s.cluster.Status.APIEndpoints) != len(s.alicloudCluster.Status.ApiEndpoints) {
		clusterPatcher, err := patch.NewHelper(s.cluster, s.client)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "patch.NewHelper error")
		}
		s.cluster.Status.APIEndpoints = s.alicloudCluster.Status.ApiEndpoints
		if err := clusterPatcher.Patch(context.TODO(), s.cluster); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "clusterPatcher.Patch error")
		}
	}

	if rs, err := s.reconcileNetwork(); err != nil {
		return rs, errors.Wrap(err, "reconcileNetwork error")
	}

	s.Info("ReconcileNormal success")
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileNetwork() (reconcile.Result, error) {
	s.Info("reconcileNetwork")
	s.alicloudCluster.Status.Message = "reconcileNetwork"

	s.alicloudCluster.Status.Message += "-reconcileVPC"
	if rs, err := s.reconcileVPC(); err != nil {
		return rs, errors.Wrap(err, "reconcileVPC")
	}
	s.alicloudCluster.Status.Message += "-reconcileVSwitch"
	if rs, err := s.reconcileVSwitch(); err != nil {
		return rs, errors.Wrap(err, "reconcileVSwitch")
	}
	s.alicloudCluster.Status.Message += "-reconcileNat"
	if rs, err := s.reconcileNat(); err != nil {
		return rs, errors.Wrap(err, "reconcileNat")
	}
	s.alicloudCluster.Status.Message += "-reconcileSLB"
	if rs, err := s.reconcileSLB(); err != nil {
		return rs, errors.Wrap(err, "reconcileSLB")
	}
	s.alicloudCluster.Status.Message += "-reconcileSecurityGroup"
	if rs, err := s.reconcileSecurityGroup(); err != nil {
		return rs, errors.Wrap(err, "reconcileSecurityGroup")
	}
	s.alicloudCluster.Status.Message += "-reconcileSSHKey"
	if rs, err := s.reconcileSSHKey(); err != nil {
		return rs, errors.Wrap(err, "reconcileSSHKey")
	}

	s.Info("reconcileNetwork success")
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileVPC() (reconcile.Result, error) {
	if len(s.alicloudCluster.Status.Network.VPC.VpcId) > 0 {
		return reconcile.Result{}, nil
	}

	s.Info("reconcileVPC")

	spec := s.alicloudCluster.Spec.Network.VPC
	id := spec.VpcId

	var err error
	var target *infrav1.VPC
	if len(id) > 0 {
		target, err = s.vpc.Describe(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "Describe %v", id)
		}
		if target == nil {
			return reconcile.Result{}, errors.Errorf("target not found: %v", id)
		}
		if target.Status != infrav1.Available {
			target, err = s.vpc.WaitReady(target.VpcId)
			if err != nil {
				return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
			}
		}
	} else {
		id, err = s.vpc.Create(spec)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "Create")
		}
		target, err = s.vpc.WaitReady(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
		}
	}

	s.Info("reconcileVPC success", "status", target)
	target.DeepCopyInto(&s.alicloudCluster.Status.Network.VPC)
	_ = s.patch()
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileVSwitch() (reconcile.Result, error) {
	if len(s.alicloudCluster.Status.Network.VSwitch.VpcId) > 0 {
		return reconcile.Result{}, nil
	}

	s.Info("reconcileVSwitch")

	spec := s.alicloudCluster.Spec.Network.VSwitch
	id := spec.VSwitchId

	var err error
	var target *infrav1.VSwitch
	if len(id) > 0 {
		target, err = s.vswitch.Describe(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "Describe %v", id)
		}
		if target == nil {
			return reconcile.Result{}, errors.Errorf("target not found: %v", id)
		}
		if target.Status != infrav1.Available {
			target, err = s.vswitch.WaitReady(target.VSwitchId)
			if err != nil {
				return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
			}
		}
	} else {
		id, err = s.vswitch.Create(spec, s.alicloudCluster.Spec.ZoneId, s.alicloudCluster.Status.Network.VPC.VpcId)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "Create")
		}
		target, err = s.vswitch.WaitReady(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
		}
	}

	s.Info("reconcileVSwitch success", "status", target)
	target.DeepCopyInto(&s.alicloudCluster.Status.Network.VSwitch)
	_ = s.patch()
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileNatGateway() (reconcile.Result, error) {
	if len(s.alicloudCluster.Status.Network.Nat.NatGateway.NatGatewayId) > 0 {
		return reconcile.Result{}, nil
	}

	s.Info("reconcileNatGateway")

	spec := s.alicloudCluster.Spec.Network.Nat.NatGateway
	id := spec.NatGatewayId

	var err error
	var target *infrav1.NatGateway
	if len(id) > 0 {
		target, err = s.vpc.DescribeNatGateway(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "Describe %v", id)
		}
		if target == nil {
			return reconcile.Result{}, errors.Errorf("target not found: %v", id)
		}
		if target.Status != infrav1.NGWAvailable {
			target, err = s.vpc.WaitNatGatewayReady(target.NatGatewayId)
			if err != nil {
				return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
			}
		}
	} else {
		ngwID, err := s.vpc.CreateNatGateway(spec, s.alicloudCluster.Status.Network.VPC.VpcId)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "Create")
		}
		target, err = s.vpc.WaitNatGatewayReady(ngwID)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
		}
	}

	s.Info("reconcileNatGateway success", "status", target)
	target.DeepCopyInto(&s.alicloudCluster.Status.Network.Nat.NatGateway)
	_ = s.patch()
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileEIP() (reconcile.Result, error) {
	if len(s.alicloudCluster.Status.Network.Nat.EIP.AllocationId) > 0 {
		return reconcile.Result{}, nil
	}

	s.Info("reconcileEIP")

	spec := s.alicloudCluster.Spec.Network.Nat.EIP
	id := spec.AllocationId

	var err error
	var target *infrav1.EIP
	if len(id) > 0 {
		target, err = s.vpc.DescribeEIP(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "Describe %v", id)
		}
		if target == nil {
			return reconcile.Result{}, errors.Errorf("target not found: %v", id)
		}
		if target.Status != infrav1.EIPAvailable && target.Status != infrav1.EIPInUse {
			target, err = s.vpc.WaitEIPStatus(target.AllocationId, infrav1.EIPAvailable, infrav1.EIPInUse)
			if err != nil {
				return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
			}
		}
	} else {
		eipID, err := s.vpc.CreateEIP(spec, s.alicloudCluster.Status.Network.VPC.VpcId)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "Create")
		}
		target, err = s.vpc.WaitEIPStatus(eipID, infrav1.EIPAvailable)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
		}
	}

	s.Info("reconcileEIP success", "status", target)
	target.DeepCopyInto(&s.alicloudCluster.Status.Network.Nat.EIP)
	_ = s.patch()
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileSSHKey() (reconcile.Result, error) {
	s.Info("reconcileSSHKey")

	ecscli, err := aliyun.GetEcsClient(s.alicloudCluster.Spec.RegionId)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "GetEcsClient")
	}

	keyreq := ecs.CreateDescribeKeyPairsRequest()
	keyreq.KeyPairName = pkg.DefaultSSHKeyName
	keyreq.RegionId = s.alicloudCluster.Spec.RegionId
	keyresp, err := ecscli.DescribeKeyPairs(keyreq)
	if err != nil || keyresp == nil {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, nil
	}

	if keyresp.TotalCount == 0 || len(keyresp.KeyPairs.KeyPair) == 0 {
		req := ecs.CreateCreateKeyPairRequest()
		req.RegionId = s.alicloudCluster.Spec.RegionId
		req.KeyPairName = pkg.DefaultSSHKeyName
		resp, err := ecscli.CreateKeyPair(req)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "CreateKeyPair")
		}
		s.Info("create default keypair", "resp", resp.KeyPairId)
	}

	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileNat() (reconcile.Result, error) {
	if len(s.alicloudCluster.Status.Network.Nat.EIP.AllocationId) > 0 && len(s.alicloudCluster.Status.Network.Nat.NatGateway.NatGatewayId) > 0 {
		return reconcile.Result{}, nil
	}

	s.Info("reconcileNat")

	s.alicloudCluster.Status.Message += "-reconcileNatGateway"
	if rs, err := s.reconcileNatGateway(); err != nil {
		return rs, errors.Wrap(err, "reconcileNatGateway")
	}
	s.alicloudCluster.Status.Message += "-reconcileEIP"
	if rs, err := s.reconcileEIP(); err != nil {
		return rs, errors.Wrap(err, "reconcileEIP")
	}

	eip := &s.alicloudCluster.Status.Network.Nat.EIP
	ngw := &s.alicloudCluster.Status.Network.Nat.NatGateway

	var err error
	s.Info("AssociateEipToNatGateway")
	s.alicloudCluster.Status.Message += "-AssociateEipToNatGateway"
	if err = s.vpc.AssociateEipToNatGateway(eip, ngw); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "AssociateEipToNatGateway")
	}

	s.Info("WaitEIPStatus")
	s.alicloudCluster.Status.Message += "-WaitEIPStatus"
	eip, err = s.vpc.WaitEIPStatus(eip.AllocationId, infrav1.EIPInUse)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "WaitEIPStatus")
	}

	s.Info("CreateSnatEntry")
	s.alicloudCluster.Status.Message += "-CreateSnatEntry"
	snatEntryId, err := s.vpc.CreateSnatEntry(eip, ngw, s.alicloudCluster.Status.Network.VSwitch.VSwitchId)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "CreateSnatEntry")
	}

	s.Info("reconcileNat success")
	s.alicloudCluster.Status.Network.Nat.SnatEntryId = snatEntryId
	_ = s.patch()
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileSLB() (reconcile.Result, error) {
	if len(s.alicloudCluster.Status.Network.SLB.LoadBalancerId) > 0 {
		return reconcile.Result{}, nil
	}

	s.Info("reconcileSLB")

	spec := s.alicloudCluster.Spec.Network.SLB
	id := spec.LoadBalancerId

	var err error
	var target *infrav1.SLB
	if len(id) > 0 {
		target, err = s.slb.Describe(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "Describe %v", id)
		}
		if target == nil {
			return reconcile.Result{}, errors.Errorf("target not found: %v", id)
		}
		if target.LoadBalancerStatus != infrav1.SLBActive {
			target, err = s.slb.WaitReady(target.LoadBalancerId)
			if err != nil {
				return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
			}
		}
	} else {
		id, err = s.slb.Create(spec, s.alicloudCluster.Status.Network.VPC.VpcId)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "Create")
		}
		target, err = s.slb.WaitReady(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
		}
	}
	target.DeepCopyInto(&s.alicloudCluster.Status.Network.SLB)

	if len(spec.VServerGroupId) > 0 {
		vsgIDs, err := s.slb.DescribeServerGroup(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "DescribeServerGroup %v", id)
		}
		var found bool
		for _, v := range vsgIDs {
			if v == spec.VServerGroupId {
				found = true
			}
		}
		if !found {
			return reconcile.Result{}, errors.Errorf("VServerGroupId not found: %v", spec.VServerGroupId)
		}
		s.alicloudCluster.Status.Network.SLB.VServerGroupId = spec.VServerGroupId
	} else {
		vsgID, err := s.slb.CreateServerGroup(spec, id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "CreateServerGroup %v", id)
		}
		s.alicloudCluster.Status.Network.SLB.VServerGroupId = vsgID
	}

	// TODO join exist slb, remove hardcode
	s.alicloudCluster.Status.Message += "-CreateTCPListener"
	if err := s.slb.CreateTCPListener(spec, id, s.alicloudCluster.Status.Network.SLB.VServerGroupId); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "CreateTCPListener %v", id)
	}

	s.alicloudCluster.Status.Message += "-StartListener"
	if err := s.slb.StartListener(id); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "StartListener %v", id)
	}

	s.alicloudCluster.Status.ApiEndpoints = []clusterv1.APIEndpoint{
		{
			Host: target.Address,
			Port: 6443,
		},
	}
	_ = s.patch()
	return reconcile.Result{}, nil
}

func (s *ClusterProcessor) reconcileSecurityGroup() (reconcile.Result, error) {
	if len(s.alicloudCluster.Status.Network.SecurityGroup.SecurityGroupId) > 0 {
		return reconcile.Result{}, nil
	}

	s.Info("reconcileSecurityGroup")

	spec := s.alicloudCluster.Spec.Network.SecurityGroup
	id := spec.SecurityGroupId

	var err error
	var target *infrav1.SecurityGroup
	if len(id) > 0 {
		target, err = s.securityGroup.Describe(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "Describe %v", id)
		}
		if target == nil {
			return reconcile.Result{}, errors.Errorf("target not found: %v", id)
		}
	} else {
		id, err = s.securityGroup.Create(spec, s.alicloudCluster.Status.Network.VPC.VpcId)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "Create")
		}
		target, err = s.securityGroup.WaitReady(id)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "WaitRead %v", id)
		}
	}

	s.Info("reconcileSecurityGroup success", "status", target)
	target.DeepCopyInto(&s.alicloudCluster.Status.Network.SecurityGroup)
	_ = s.patch()
	return reconcile.Result{}, nil
}
