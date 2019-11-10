// +kubebuilder:validation:Optional

package v1alpha2

const (
	ClusterFinalizer = "alicloud-cluster.infrastructure.cluster.x-k8s.io"

	Pending   = "Pending"
	Available = "Available"

	SLBInactive = "inactive"
	SLBActive   = "active"
	SLBLocked   = "locked"

	EIPAssociating   = "Associating"
	EIPUnassociating = "Unassociating"
	EIPInUse         = "InUse"
	EIPAvailable     = "Available"

	NGWInitiating = "Initiating"
	NGWAvailable  = "Available"
	NGWPending    = "Pending"
)

type SLBStatus string

type NetworkSpec struct {
	VPC           VPCSpec           `json:"vpc,omitempty"`
	VSwitch       VSwitchSpec       `json:"vSwitch,omitempty"`
	Nat           NatSpec           `json:"nat,omitempty"`
	SLB           SLBSpec           `json:"slb,omitempty"`
	SecurityGroup SecurityGroupSpec `json:"securityGroup,omitempty"`
}

// VPCSpec 专有网络
// 使用云资源前, 必须先创建一个专有网络和交换机
// 详细文档见 [CreateVpc](https://help.aliyun.com/document_detail/35737.html)
type VPCSpec struct {
	// 使用一个已经存在的VPC
	VpcId string `json:"vpcId,omitempty"`

	// 专有网络名称。长度为2-128个字符，必须以字母或中文开头，可包含数字，点号（.），下划线（_）和短横线（-），但不能以http://或https://开头。
	VpcName string `json:"vpcName,omitempty"`
	// VPC的网段。您可以使用以下网段或其子集：
	//   10.0.0.0/8。
	//   172.16.0.0/12（默认值）。
	//   192.168.0.0/16。
	CidrBlock string `json:"cidrBlock,omitempty"`

	// VPC的描述信息。长度为2-256个字符，必须以字母或中文开头，但不能以http://或https://开头。
	Description string `json:"description,omitempty"`

	//// 用户侧网络的网段，如需定义多个网段请使用半角逗号隔开，最多支持3个网段。
	////
	//// VPC定义的默认私网转发网段为10.0.0.0/8、172.16.0.0/12、192.168.0.0/16、100.64.0.0/10和VPC CIDR网段。
	//// 如果ECS实例或弹性网卡已经具备了公网访问能力（ECS实例分配了固定公网IP、ECS实例或弹性网卡绑定了公网IP、ECS实例或弹性网卡设置了DNAT IP映射规则），
	//// 这类资源访问非上述默认私网转发网段的请求均会通过公网IP直接转发至公网。
	//// 当希望按照路由表在私网（如VPC内、通过VPN/高速通道/云企业网搭建的混合云网络）转发访问非上述默认私网网段的请求时，
	//// 需要将网络请求的目的网段设置为ECS或弹性网卡所在VPC的UserCidr。为VPC设置UserCidr后，
	//// 该VPC中访问UserCidr地址的请求将按照路由表进行转发，而不通过公网IP转发。
	//UserCidr string `json:"userCidr,omitempty"`
	//// 是否开启IPv6网段，取值：
	////   false（默认值）：不开启。
	////   true：开启。
	//EnableIpv6 string `json:"enableIpv6,omitempty"`
	//// VPC的IPv6网段
	//Ipv6CidrBlock string `json:"ipv6CidrBlock,omitempty"`
}

// VSwitchSpec 交换机,
// 使用云资源前, 必须先创建一个专有网络和交换机
// 详细文档见 [CreateVSwitch](https://help.aliyun.com/document_detail/35745.html)
type VSwitchSpec struct {
	// 使用一个已经存在的VSwitch
	VSwitchId string `json:"vSwitchId,omitempty"`

	// 交换机的名称。
	//   长度为 2-128个字符，必须以字母或中文开头，但不能以http://或https://开头。
	VSwitchName string `json:"vSwitchName,omitempty"`
	// 交换机的网段。交换机网段要求如下：
	//   交换机网段的掩码长度范围为16-29位。
	//   交换机的网段必须从属于所在VPC的网段。
	//   交换机的网段不能与所在VPC中路由条目的目标网段相同，但可以是目标网段的子集。
	//   如果交换机的网段与所在VPC的网段相同时，VPC只能有一个交换机。
	CidrBlock string `json:"cidrBlock,omitempty"`

	// 交换机的描述信息。
	//   长度为 2-256个字符，必须以字母或中文开头，但不能以http:// 或https://开头。
	Description string `json:"description,omitempty"`

	//// 交换机的IPv6网段，支持自定义VPC IPv6网段的最后8bit。取值：0-255（十进制）。
	////   交换机的IPv6网段掩码默认为64位。
	//Ipv6CidrBlock string `json:"ipv6CidrBlock,omitempty"`
}

// NatSpec NAT网关相关配置, 在VPC环境下构建一个公网流量的出入口
type NatSpec struct {
	// NAT网
	NatGateway NatGatewaySpec `json:"natGateway,omitempty"`
	//
	EIP EIPSpec `json:"eip,omitempty"`
}

// NatGatewaySpec NAT网关 在VPC环境下构建一个公网流量的出入口
// 详细文档见 [CreateNatGateway](https://help.aliyun.com/document_detail/36048.html)
type NatGatewaySpec struct {
	// 使用一个已经存在的NAT网关
	NatGatewayId string `json:"natGatewayId,omitempty"`

	// NAT网关的名称。
	//   名称在2~128个字符之间，必须以英文字母或中文开头，不能以http://和https://开头，可包含数字、点号（.）、下划线（_）或短横线（-）。
	//   如果没有指定该参数，默认使用网关ID。
	Name string `json:"name,omitempty"`

	// NAT网关的描述。
	//   描述在2~256个字符之间，不能以http://和https://开头。
	Description string `json:"description,omitempty"`
	// NAT网关的规格。取值：
	//   Small(默认值)：小型
	//   Middle：中型
	//   Large：大型
	//   XLarge.1：超大型
	Spec string `json:"spec,omitempty"`
	// 购买时长。
	//   当PricingCycle取值Month时，Period取值范围为1~9。
	//   当PricingCycle取值Year时，Period取值范围为1~3。
	//   如果InstanceChargeType参数的值为PrePaid时，该参数必选。
	Duration string `json:"duration,omitempty"`
	// 计费方式，取值：
	//   PrePaid：包年包月。
	//   PostPaid（默认值）：按量计费。
	InstanceChargeType string `json:"instanceChargeType,omitempty"`
	// 是否自动付费，取值：
	//   false：不开启自动付费，生成订单后需要到订单中心完成支付。
	//   true：开启自动付费，自动支付订单。
	// 当InstanceChargeType参数的值为PrePaid时，该参数必选；当InstanceChargeType参数的值为PostPaid时，该参数可不填。
	AutoPay string `json:"autoPay,omitempty"`
	// 包年包月的计费周期，取值：
	//   Month（默认值）：按月付费。
	//   Year：按年付费。
	//当InstanceChargeType参数的值为PrePaid时，该参数必选；当InstanceChargeType参数的值为PostPaid时，该参数可不填。
	PricingCycle string `json:"pricingCycle,omitempty"`
}

// EIPSpec 弹性公网IP 配置DNAT或SNAT功能前，需要为已创建的NAT网关绑定弹性公网IP
// 详细文档见 [AllocateEipAddress](https://help.aliyun.com/document_detail/36016.html)
type EIPSpec struct {
	// 使用一个已经存在的弹性公网IP
	AllocationId string `json:"allocationId,omitempty"`

	// EIP的带宽峰值，单位为Mbps，默认值为5。
	Bandwidth string `json:"bandwidth,omitempty"`

	// 线路类型，默认值为BGP。
	//   对于已开通单线带宽白名单的用户，ISP字段可以设置为ChinaTelecom、ChinaUnicom和ChinaMobile，用来开通中国电信、中国联通、中国移动的单线EIP。
	//   如果是杭州金融云用户，该字段必填，取值：BGP_FinanceCloud。
	ISP string `json:"isp,omitempty"`
	// EIP的计费方式，取值：
	//   PrePaid：包年包月。
	//   PostPaid（默认值）：按量计费。
	//
	//   当InstanceChargeType取值为PrePaid时，InternetChargeType必须取值PayByBandwidth；当InstanceChargeType取值为PostPaid时，InternetChargeType可取值PayByBandwidth或PayByTraffic。
	//   包年包月和按量计费的详细信息，请参见包年包月和按量计费。
	InstanceChargeType string `json:"instanceChargeType,omitempty"`
	// EIP的计量方式，取值：
	//   PayByBandwidth（默认值）：按带宽计费。
	//   PayByTraffic：按流量计费。
	//
	//   当InstanceChargeType取值为PrePaid时，InternetChargeType必须取值PayByBandwidth。详细信息，请参见包年包月。
	//   当InstanceChargeType取值为PostPaid时，InternetChargeType可取值PayByBandwidth或PayByTraffic。详细信息，请参见按使用流量和按固定带宽。
	InternetChargeType string `json:"internetChargeType,omitempty"`
	// 包年包月的计费周期，取值：
	//   Month（默认值）：按月付费。
	//   Year：按年付费。
	// 当InstanceChargeType参数的值为PrePaid时，该参数必选；当InstanceChargeType参数的值为PostPaid时，该参数可不填。
	PricingCycle string `json:"pricingCycle,omitempty"`
	// 购买时长。
	//   当PricingCycle取值Month时，Period取值范围为1~9。
	//   当PricingCycle取值Year时，Period取值范围为1~3。
	//   如果InstanceChargeType参数的值为PrePaid时，该参数必选。
	Period string `json:"period,omitempty"`
	// 是否自动付费，取值：
	//   false：不开启自动付费，生成订单后需要到订单中心完成支付。
	//   true：开启自动付费，自动支付订单。
	// 当InstanceChargeType参数的值为PrePaid时，该参数必选；当InstanceChargeType参数的值为PostPaid时，该参数可不填。
	AutoPay string `json:"autoPay,omitempty"`
}

// SLBSpec 负载均衡（Server Load Balancer）是对多台云服务器进行流量分发的负载均衡服务,
// 流量分发到apiserver
// 详细文档见 https://help.aliyun.com/document_detail/27566.html
type SLBSpec struct {
	// 使用一个已经存在的负载均衡
	LoadBalancerId string `json:"loadBalancerId,omitempty"`
	// 使用一个已经存在的后端服务器组
	VServerGroupId string `json:"vServerGroupId,omitempty"`

	// 负载均衡实例的名称。
	//   长度为2-128个英文或中文字符，必须以大小字母或中文开头，可包含数字，点号（.），下划线（_）和短横线（-），字段长度不能超过80。
	//   不指定该参数时，默认由系统分配一个实例名称。
	LoadBalancerName string `json:"loadBalancerName,omitempty"`
	// 负载均衡实例的网络类型。取值：
	//   internet：创建公网负载均衡实例后，系统会分配一个公网IP地址，可以转发公网请求。
	//   intranet：创建内网负载均衡实例后，系统会分配一个内网IP地址，仅可转发内网请求。
	AddressType string `json:"addressType,omitempty"`
	// 监听的带宽峰值
	Bandwidth string `json:"bandwidth,omitempty"`
	// 负载均衡实例的IP版本，可以设置为ipv4或者ipv6
	AddressIPVersion string `json:"addressIPVersion,omitempty"`
	// 后端服务器组名
	VServerGroupName string `json:"vServerGroupName,omitempty"`

	// 指定负载均衡实例的私网IP地址，该地址必须包含在交换机的目标网段下。
	Address string `json:"address,omitempty"`
	// 负载均衡实例的规格。取值： https://help.aliyun.com/document_detail/85931.html?spm=a2c1g.8271268.0.0.2e90df253aqA3R
	LoadBalancerSpec string `json:"loadBalancerSpec,omitempty"`
	CloudType        string `json:"cloudType,omitempty"`
	// 负载均衡实例的主可用区ID。
	MasterZoneId string `json:"masterZoneId,omitempty"`
	//预付费公网实例的购买时长，取值：
	//  如果PricingCycle为month，取值为1~9。
	//  如果PricingCycle为year，取值为1~3。
	//  该参数仅适用于中国站。
	// 负载均衡实例的备可用区ID。
	SlaveZoneId string `json:"slaveZoneId,omitempty"`
	// 是否开启实例删除保护
	DeleteProtection string `json:"deleteProtection,omitempty"`
	// 公网类型实例的付费方式。取值：
	//   paybybandwidth：按带宽计费。
	//   paybytraffic：按流量计费（默认值）。
	InternetChargeType string `json:"internetChargeType,omitempty"`
	// 实例的计费类型，取值：
	//   PayOnDemand：按量付费。
	//   PrePay：预付费。
	PayType string `json:"payType,omitempty"`
	//是否是自动支付预付费公网实例的账单。
	//  取值：true|false（默认）。
	//  该参数仅适用于中国站。
	AutoPay string `json:"autoPay,omitempty"`
	// 预付费公网实例的计费周期，取值：month|year
	// 仅适用于中国站。
	PricingCycle string `json:"pricingCycle,omitempty"`
}

// SecurityGroupSpec 安全组, 在创建ECS实例时必须指定安全组，每台ECS实例至少属于一个安全组
// 详细文档见 [CreateSecurityGroup](https://help.aliyun.com/document_detail/25553.html)
type SecurityGroupSpec struct {
	// 使用一个已经存在的安全组
	SecurityGroupId string `json:"securityGroupId,omitempty"`

	// 安全组名称。长度为2~128个英文或中文字符。必须以大小字母或中文开头，不能以 http://和https://开头。可以包含数字、半角冒号（:）、下划线（_）或者连字符（-）。默认值：空。
	SecurityGroupName string `json:"securityGroupName,omitempty"`
	// 安全组入方向规则
	Rules []*SecurityGroupRuleSpec `json:"rules,omitempty"`

	// 安全组描述信息。长度为2~256个英文或中文字符，不能以 http://和https://开头。 默认值：空。
	Description string `json:"description,omitempty"`
	// 安全组类型，分为普通安全组与企业安全组。取值范围：
	//   normal：普通安全组。
	//   enterprise：企业安全组。https://help.aliyun.com/document_detail/120621.html?spm=a2c1g.8271268.0.0.2e90df253aqA3R
	SecurityGroupType string `json:"securityGroupType,omitempty"`
}

// SecurityGroupRuleSpec 安全组入方向规则
// 详细文档见 [AuthorizeSecurityGroup](https://help.aliyun.com/document_detail/25554.html)
type SecurityGroupRuleSpec struct {
	// 网卡类型。取值范围：
	//   internet：公网网卡。
	//   intranet：内网网卡。
	//   当设置安全组之间互相访问时，即指定了DestGroupId且没有指定DestCidrIp时，参数NicType取值只能为intranet。
	//   默认值：internet。
	NicType string `json:"nicType,omitempty"`
	// 传输层协议。不区分大小写。取值范围：
	//   icmp
	//   gre
	//   tcp
	//   udp
	//   all：支持所有协议
	IpProtocol string `json:"ipProtocol,omitempty"`
	// 源端IP地址范围。支持CIDR格式和IPv4格式的IP地址范围。
	//   默认值：0.0.0.0/0。
	SourceCidrIp string `json:"sourceCidrIp,omitempty"`
	// 目的端安全组开放的传输层协议相关的端口范围。取值范围：
	//   TCP/UDP协议：取值范围为1~65535。使用斜线（/）隔开起始端口和终止端口。正确示范：1/200；错误示范：200/1。
	//   ICMP协议：-1/-1。
	//   GRE协议：-1/-1。
	//   all：-1/-1。
	PortRange string `json:"portRange,omitempty"`

	// 安全组规则的描述信息。长度为1~512个字符
	Description string `json:"description,omitempty"`
	// 设置访问权限的源端安全组ID。必须设置SourceGroupId或者SourceCidrIp参数。
	//   如果指定了SourceGroupId没有指定参数SourceCidrIp，则参数NicType取值只能为 intranet。
	//   如果同时指定了SourceGroupId和SourceCidrIp，则默认以SourceCidrIp为准。
	SourceGroupId string `json:"sourceGroupId,omitempty"`
	// 跨账户设置安全组规则时，源端安全组所属的阿里云账户ID。
	//   如果SourceGroupOwnerId及SourceGroupOwnerAccount均未设置，则认为是设置您其他安全组的访问权限。
	//   如果您已经设置参数SourceCidrIp，则参数SourceGroupOwnerId无效。
	SourceGroupOwnerId string `json:"sourceGroupOwnerId,omitempty"`
	// 跨账户设置安全组规则时，源端安全组所属的阿里云账户。
	//   如果SourceGroupOwnerAccount及SourceGroupOwnerID均未设置，则认为是设置您其他安全组的访问权限。
	//   如果已经设置参数SourceCidrIp，则参数SourceGroupOwnerAccount无效。
	SourceGroupOwnerAccount string `json:"sourceGroupOwnerAccount,omitempty"`
	// 安全组规则优先级。取值范围：1~100
	//   默认值：1。
	Priority string `json:"priority,omitempty"`
	// 访问权限。取值范围：
	//   accept：接受访问。
	//   drop：拒绝访问，不返回拒绝信息。
	//   默认值：accept。
	Policy string `json:"policy,omitempty"`
	// 源端IPv6 CIDR地址段。支持CIDR格式和IPv6格式的IP地址范围。
	//   仅支持VPC类型的IP地址。
	//   默认值：无。
	Ipv6SourceCidrIp string `json:"ipv6SourceCidrIp,omitempty"`
	// 源端安全组开放的传输层协议相关的端口范围。取值范围：
	//   TCP/UDP协议：取值范围为1~65535。使用斜线（/）隔开起始端口和终止端口。正确示范：1/200；错误示范：200/1。
	//   ICMP协议：-1/-1。
	//   GRE协议：-1/-1。
	//   all：-1/-1。
	SourcePortRange string `json:"sourcePortRange,omitempty"`
	// 目的端IP地址范围。支持CIDR格式和IPv4格式的IP地址范围。
	//   默认值：0.0.0.0/0。
	DestCidrIp string `json:"destCidrIp,omitempty"`
	// 目的端IPv6 CIDR地址段。支持CIDR格式和IPv6格式的IP地址范围。
	//   仅支持VPC类型的IP地址。
	//   默认值：无。
	Ipv6DestCidrIp string `json:"ipv6DestCidrIp,omitempty"`
}

///////////////////////////////

type Network struct {
	VPC           VPC           `json:"vpc,omitempty"`
	VSwitch       VSwitch       `json:"vSwitch,omitempty"`
	SLB           SLB           `json:"slb,omitempty"`
	Nat           Nat           `json:"nat,omitempty"`
	SecurityGroup SecurityGroup `json:"securityGroup,omitempty"`
}

type VPC struct {
	VpcId           string `json:"vpcId,omitempty"`
	RegionId        string `json:"regionId,omitempty"`
	Status          string `json:"status,omitempty"`
	VpcName         string `json:"vpcName,omitempty"`
	CreationTime    string `json:"creationTime,omitempty"`
	CidrBlock       string `json:"cidrBlock,omitempty"`
	Ipv6CidrBlock   string `json:"ipv6CidrBlock,omitempty"`
	VRouterId       string `json:"vRouterId,omitempty"`
	Description     string `json:"description,omitempty"`
	IsDefault       bool   `json:"isDefault,omitempty"`
	NetworkAclNum   string `json:"networkAclNum,omitempty"`
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	CenStatus       string `json:"cenStatus,omitempty"`
}

type VSwitch struct {
	VSwitchId               string `json:"vSwitchId,omitempty"`
	VpcId                   string `json:"vpcId,omitempty"`
	Status                  string `json:"status,omitempty"`
	CidrBlock               string `json:"cidrBlock,omitempty"`
	Ipv6CidrBlock           string `json:"ipv6CidrBlock,omitempty"`
	ZoneId                  string `json:"zoneId,omitempty"`
	AvailableIpAddressCount int64  `json:"availableIpAddressCount,omitempty"`
	Description             string `json:"description,omitempty"`
	VSwitchName             string `json:"vSwitchName,omitempty"`
	CreationTime            string `json:"creationTime,omitempty"`
	IsDefault               bool   `json:"isDefault,omitempty"`
	ResourceGroupId         string `json:"resourceGroupId,omitempty"`
	NetworkAclId            string `json:"networkAclId,omitempty"`
}

type Nat struct {
	NatGateway  NatGateway `json:"natGateway,omitempty"`
	EIP         EIP        `json:"eip,omitempty"`
	SnatEntryId string     `json:"snatEntryId,omitempty"`
}

type NatGateway struct {
	NatGatewayId       string                            `json:"natGatewayId,omitempty"`
	Name               string                            `json:"name,omitempty"`
	Description        string                            `json:"description,omitempty"`
	VpcId              string                            `json:"vpcId,omitempty"`
	Spec               string                            `json:"spec,omitempty"`
	InstanceChargeType string                            `json:"instanceChargeType,omitempty"`
	ExpiredTime        string                            `json:"expiredTime,omitempty"`
	AutoPay            bool                              `json:"autoPay,omitempty"`
	BusinessStatus     string                            `json:"businessStatus,omitempty"`
	CreationTime       string                            `json:"creationTime,omitempty"`
	Status             string                            `json:"status,omitempty"`
	DeletionProtection bool                              `json:"deletionProtection,omitempty"`
	SnatTableIds       SnatTableIdsInDescribeNatGateways `json:"snatTableIds,omitempty"`
}

type SnatTableIdsInDescribeNatGateways struct {
	SnatTableId []string `json:"SnatTableId" xml:"SnatTableId"`
}

type EIP struct {
	IpAddress          string `json:"ipAddress,omitempty"`
	PrivateIpAddress   string `json:"privateIpAddress,omitempty"`
	AllocationId       string `json:"allocationId,omitempty"`
	Status             string `json:"status,omitempty"`
	InstanceId         string `json:"instanceId,omitempty"`
	Bandwidth          string `json:"bandwidth,omitempty"`
	EipBandwidth       string `json:"eipBandwidth,omitempty"`
	InternetChargeType string `json:"internetChargeType,omitempty"`
	AllocationTime     string `json:"allocationTime,omitempty"`
	InstanceType       string `json:"instanceType,omitempty"`
	InstanceRegionId   string `json:"instanceRegionId,omitempty"`
	ChargeType         string `json:"chargeType,omitempty"`
	ExpiredTime        string `json:"expiredTime,omitempty"`
	HDMonitorStatus    string `json:"hdMonitorStatus,omitempty"`
	Name               string `json:"name,omitempty"`
	ISP                string `json:"isp,omitempty"`
	Descritpion        string `json:"descritpion,omitempty"`
	ResourceGroupId    string `json:"resourceGroupId,omitempty"`
	HasReservationData string `json:"hasReservationData,omitempty"`
	Mode               string `json:"mode,omitempty"`
	DeletionProtection bool   `json:"deletionProtection,omitempty"`
	SecondLimited      bool   `json:"secondLimited,omitempty"`
}

type SLB struct {
	LoadBalancerId     string `json:"loadBalancerId,omitempty"`
	LoadBalancerName   string `json:"loadBalancerName,omitempty"`
	LoadBalancerStatus string `json:"loadBalancerStatus,omitempty"`
	Address            string `json:"address,omitempty"`
	AddressType        string `json:"addressType,omitempty"`
	RegionId           string `json:"regionId,omitempty"`
	RegionIdAlias      string `json:"regionIdAlias,omitempty"`
	VSwitchId          string `json:"vSwitchId,omitempty"`
	VpcId              string `json:"vpcId,omitempty"`
	NetworkType        string `json:"networkType,omitempty"`
	MasterZoneId       string `json:"masterZoneId,omitempty"`
	SlaveZoneId        string `json:"slaveZoneId,omitempty"`
	InternetChargeType string `json:"internetChargeType,omitempty"`
	CreateTime         string `json:"createTime,omitempty"`
	CreateTimeStamp    int64  `json:"createTimeStamp,omitempty"`
	PayType            string `json:"payType,omitempty"`
	ResourceGroupId    string `json:"resourceGroupId,omitempty"`
	AddressIPVersion   string `json:"addressIPVersion,omitempty"`
	VServerGroupId     string `json:"vServerGroupId,omitempty"`
}

type SecurityGroup struct {
	SecurityGroupId         string `json:"securityGroupId,omitempty"`
	Description             string `json:"description,omitempty"`
	SecurityGroupName       string `json:"securityGroupName,omitempty"`
	VpcId                   string `json:"vpcId,omitempty"`
	CreationTime            string `json:"creationTime,omitempty"`
	SecurityGroupType       string `json:"securityGroupType,omitempty"`
	AvailableInstanceAmount int    `json:"availableInstanceAmount,omitempty"`
	EcsCount                int    `json:"ecsCount,omitempty"`
	ResourceGroupId         string `json:"resourceGroupId,omitempty"`
}
