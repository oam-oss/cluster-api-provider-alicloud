package v1alpha2

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"k8s.io/apimachinery/pkg/util/rand"
)

func (s *VPCSpec) ConvertToCreateReq() *vpc.CreateVpcRequest {
	req := vpc.CreateCreateVpcRequest()
	req.Scheme = "https"
	req.ClientToken = rand.String(32)

	req.Description = s.Description
	req.VpcName = s.VpcName
	req.CidrBlock = s.CidrBlock

	return req
}

func (s *VSwitchSpec) ConvertToCreateReq(vpcId string, zoneID string) *vpc.CreateVSwitchRequest {
	req := vpc.CreateCreateVSwitchRequest()
	req.Scheme = "https"
	req.ClientToken = rand.String(32)

	req.Description = s.Description
	req.VpcId = vpcId
	req.VSwitchName = s.VSwitchName
	req.CidrBlock = s.CidrBlock
	req.ZoneId = zoneID

	return req
}

func (s *NatGatewaySpec) ConvertToCreateReq(vpcId string) *vpc.CreateNatGatewayRequest {
	req := vpc.CreateCreateNatGatewayRequest()
	req.Scheme = "https"
	req.ClientToken = rand.String(32)

	req.VpcId = vpcId

	req.Description = s.Description
	req.Spec = s.Spec
	req.Duration = s.Duration
	req.InstanceChargeType = s.InstanceChargeType
	req.AutoPay = requests.Boolean(s.AutoPay)
	req.Name = s.Name
	req.PricingCycle = s.PricingCycle

	return req
}
func (s *EIPSpec) ConvertToCreateReq(vpcId string) *vpc.AllocateEipAddressRequest {
	req := vpc.CreateAllocateEipAddressRequest()
	req.Scheme = "https"
	req.ClientToken = rand.String(32)

	req.ISP = s.ISP
	req.InstanceChargeType = s.InstanceChargeType
	req.Period = requests.Integer(s.Period)
	req.AutoPay = requests.Boolean(s.AutoPay)
	req.Bandwidth = s.Bandwidth
	req.InternetChargeType = s.InternetChargeType
	req.PricingCycle = s.PricingCycle

	return req
}

func (s *SLBSpec) ConvertToCreateReq(vpcId string) *slb.CreateLoadBalancerRequest {
	req := slb.CreateCreateLoadBalancerRequest()
	req.Scheme = "https"
	req.ClientToken = rand.String(32)

	req.VpcId = vpcId

	req.LoadBalancerName = s.LoadBalancerName
	req.AddressType = s.AddressType
	req.Address = s.Address
	req.Bandwidth = requests.Integer(s.Bandwidth)
	req.AddressIPVersion = s.AddressIPVersion
	req.LoadBalancerSpec = s.LoadBalancerSpec
	req.CloudType = s.CloudType
	req.MasterZoneId = s.MasterZoneId
	req.SlaveZoneId = s.SlaveZoneId
	req.DeleteProtection = s.DeleteProtection

	req.InternetChargeType = s.InternetChargeType
	req.PayType = s.PayType
	req.AutoPay = requests.Boolean(s.AutoPay)
	req.PricingCycle = s.PricingCycle

	return req
}

func (s *SLBSpec) ConvertToCreateSLBVGReq(slbID string) *slb.CreateVServerGroupRequest {
	req := slb.CreateCreateVServerGroupRequest()
	req.Scheme = "https"

	req.LoadBalancerId = slbID
	req.VServerGroupName = s.VServerGroupName

	return req
}

func (s *SLBSpec) ConvertToCreateSLBTCPListenerReq(slbID string, vgID string) *slb.CreateLoadBalancerTCPListenerRequest {
	req := slb.CreateCreateLoadBalancerTCPListenerRequest()
	req.Scheme = "https"

	req.LoadBalancerId = slbID
	req.VServerGroupId = vgID
	req.Bandwidth = "100"
	req.ListenerPort = requests.NewInteger(6443)
	req.BackendServerPort = requests.NewInteger(6443)

	return req
}

func (s *SLBSpec) ConvertToStartSLBLisenerReq(slbID string) *slb.StartLoadBalancerListenerRequest {
	req := slb.CreateStartLoadBalancerListenerRequest()
	req.Scheme = "https"

	req.LoadBalancerId = slbID
	req.ListenerPort = requests.NewInteger(6443)

	return req
}

func (s *SecurityGroupSpec) ConvertToCreateReq(vpcId string) *ecs.CreateSecurityGroupRequest {
	req := ecs.CreateCreateSecurityGroupRequest()
	req.Scheme = "https"
	req.ClientToken = rand.String(32)

	req.Description = s.Description
	req.SecurityGroupName = s.SecurityGroupName
	req.SecurityGroupType = s.SecurityGroupType
	req.VpcId = vpcId

	return req
}

func (s *SecurityGroupSpec) ConvertToAuthorizeSecurityGroupRequest(sgID string) []*ecs.AuthorizeSecurityGroupRequest {
	var list []*ecs.AuthorizeSecurityGroupRequest
	for _, r := range s.Rules {
		req := ecs.CreateAuthorizeSecurityGroupRequest()
		req.Scheme = "https"
		req.ClientToken = rand.String(32)
		req.SecurityGroupId = sgID

		req.NicType = r.NicType
		req.SourcePortRange = r.SourcePortRange
		req.Description = r.Description
		req.SourceGroupOwnerId = requests.Integer(r.SourceGroupOwnerId)
		req.SourceGroupOwnerAccount = r.SourceGroupOwnerAccount
		req.Ipv6SourceCidrIp = r.Ipv6SourceCidrIp
		req.Ipv6DestCidrIp = r.Ipv6DestCidrIp
		req.Policy = r.Policy
		req.PortRange = r.PortRange
		req.IpProtocol = r.IpProtocol
		req.SourceCidrIp = r.SourceCidrIp
		req.Priority = r.Priority
		req.DestCidrIp = r.DestCidrIp
		req.SourceGroupId = r.SourceGroupId

		list = append(list, req)
	}

	return list
}

func (s *SLB) FillFrom(desc *slb.LoadBalancer) {
	s.LoadBalancerId = desc.LoadBalancerId
	s.LoadBalancerName = desc.LoadBalancerName
	s.LoadBalancerStatus = desc.LoadBalancerStatus
	s.Address = desc.Address
	s.AddressType = desc.AddressType
	s.RegionId = desc.RegionId
	s.RegionIdAlias = desc.RegionIdAlias
	s.VSwitchId = desc.VSwitchId
	s.VpcId = desc.VpcId
	s.NetworkType = desc.NetworkType
	s.MasterZoneId = desc.MasterZoneId
	s.SlaveZoneId = desc.SlaveZoneId
	s.InternetChargeType = desc.InternetChargeType
	s.CreateTime = desc.CreateTime
	s.CreateTimeStamp = desc.CreateTimeStamp
	s.PayType = desc.PayType
	s.ResourceGroupId = desc.ResourceGroupId
	s.AddressIPVersion = desc.AddressIPVersion
}

func (s *VPC) FillFrom(desc *vpc.Vpc) {
	s.VpcId = desc.VpcId
	s.RegionId = desc.RegionId
	s.Status = desc.Status
	s.VpcName = desc.VpcName
	s.CreationTime = desc.CreationTime
	s.CidrBlock = desc.CidrBlock
	s.Ipv6CidrBlock = desc.Ipv6CidrBlock
	s.VRouterId = desc.VRouterId
	s.Description = desc.Description
	s.IsDefault = desc.IsDefault
	s.NetworkAclNum = desc.NetworkAclNum
	s.ResourceGroupId = desc.ResourceGroupId
	s.CenStatus = desc.CenStatus
}

func (s *VSwitch) FillFrom(desc *vpc.VSwitch) {
	s.VSwitchId = desc.VSwitchId
	s.VpcId = desc.VpcId
	s.Status = desc.Status
	s.CidrBlock = desc.CidrBlock
	s.ZoneId = desc.ZoneId
	s.AvailableIpAddressCount = desc.AvailableIpAddressCount
	s.Description = desc.Description
	s.VSwitchName = desc.VSwitchName
	s.CreationTime = desc.CreationTime
	s.IsDefault = desc.IsDefault
	s.ResourceGroupId = desc.ResourceGroupId
	s.NetworkAclId = desc.NetworkAclId
}

func (s *NatGateway) FillFrom(resp *vpc.NatGateway) {
	s.NatGatewayId = resp.NatGatewayId
	s.Name = resp.Name
	s.Description = resp.Description
	s.VpcId = resp.VpcId
	s.Spec = resp.Spec
	s.InstanceChargeType = resp.InstanceChargeType
	s.ExpiredTime = resp.ExpiredTime
	s.AutoPay = resp.AutoPay
	s.BusinessStatus = resp.BusinessStatus
	s.CreationTime = resp.CreationTime
	s.Status = resp.Status
	s.DeletionProtection = resp.DeletionProtection
	for _, i := range resp.SnatTableIds.SnatTableId {
		s.SnatTableIds.SnatTableId = append(s.SnatTableIds.SnatTableId, i)
	}
}

func (s *EIP) FillFrom(resp *vpc.EipAddress) {
	s.IpAddress = resp.IpAddress
	s.PrivateIpAddress = resp.PrivateIpAddress
	s.AllocationId = resp.AllocationId
	s.Status = resp.Status
	s.InstanceId = resp.InstanceId
	s.Bandwidth = resp.Bandwidth
	s.EipBandwidth = resp.EipBandwidth
	s.InternetChargeType = resp.InternetChargeType
	s.AllocationTime = resp.AllocationTime
	s.InstanceType = resp.InstanceType
	s.InstanceRegionId = resp.InstanceRegionId
	s.ChargeType = resp.ChargeType
	s.ExpiredTime = resp.ExpiredTime
	s.HDMonitorStatus = resp.HDMonitorStatus
	s.Name = resp.Name
	s.ISP = resp.ISP
	s.Descritpion = resp.Descritpion
	s.ResourceGroupId = resp.ResourceGroupId
	s.HasReservationData = resp.HasReservationData
	s.Mode = resp.Mode
	s.DeletionProtection = resp.DeletionProtection
	s.SecondLimited = resp.SecondLimited
}

func (s *SecurityGroup) FillFrom(desc *ecs.SecurityGroup) {
	s.SecurityGroupId = desc.SecurityGroupId
	s.Description = desc.Description
	s.SecurityGroupName = desc.SecurityGroupName
	s.VpcId = desc.VpcId
	s.CreationTime = desc.CreationTime
	s.SecurityGroupType = desc.SecurityGroupType
	s.AvailableInstanceAmount = desc.AvailableInstanceAmount
	s.EcsCount = desc.EcsCount
	s.ResourceGroupId = desc.ResourceGroupId
}

func InstanceFromEcs(instance *ecs.Instance) *Instance {
	if instance == nil {
		return nil
	}
	return &Instance{
		ImageId:                 instance.ImageId,
		InstanceType:            instance.InstanceType,
		OsType:                  instance.OsType,
		DeviceAvailable:         instance.DeviceAvailable,
		InstanceNetworkType:     instance.InstanceNetworkType,
		LocalStorageAmount:      instance.LocalStorageAmount,
		NetworkType:             instance.NetworkType,
		InstanceChargeType:      instance.InstanceChargeType,
		InstanceName:            instance.InstanceName,
		StartTime:               instance.StartTime,
		ZoneId:                  instance.ZoneId,
		InternetChargeType:      instance.InternetChargeType,
		InternetMaxBandwidthIn:  instance.InternetMaxBandwidthIn,
		HostName:                instance.HostName,
		Status:                  instance.Status,
		CPU:                     instance.CPU,
		Cpu:                     instance.Cpu,
		OSName:                  instance.OSName,
		OSNameEn:                instance.OSNameEn,
		SerialNumber:            instance.SerialNumber,
		RegionId:                instance.RegionId,
		InternetMaxBandwidthOut: instance.InternetMaxBandwidthOut,
		InstanceTypeFamily:      instance.InstanceTypeFamily,
		InstanceId:              instance.InstanceId,
		Description:             instance.Description,
		ExpiredTime:             instance.ExpiredTime,
		OSType:                  instance.OSType,
		Memory:                  instance.Memory,
		CreationTime:            instance.CreationTime,
		KeyPairName:             instance.KeyPairName,
		LocalStorageCapacity:    instance.LocalStorageCapacity,
		VlanId:                  instance.VlanId,
		StoppedMode:             instance.StoppedMode,
	}
}
