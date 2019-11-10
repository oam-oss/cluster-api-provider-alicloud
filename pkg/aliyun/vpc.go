package aliyun

import (
	"fmt"
	"strings"

	sdkerr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg/aliyun/retry"
)

func NewVPCClient(logger logr.Logger, regionID string) (*VPCClient, error) {
	cli, err := vpc.NewClientWithAccessKey(regionID, AccessKeyId, AccessKeySecret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vpc client")
	}
	return &VPCClient{
		Logger: logger.WithValues("client", "vpc"),
		cli:    cli,
	}, nil
}

type VPCClient struct {
	logr.Logger
	cli *vpc.Client
}

func (s *VPCClient) Describe(id string) (*infrav1.VPC, error) {
	logger := s.WithValues("SDKAction", "Describe", "id", id)

	req := vpc.CreateDescribeVpcsRequest()
	req.Scheme = "https"
	req.VpcId = id

	var resp *vpc.DescribeVpcsResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		resp, err = s.cli.DescribeVpcs(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DescribeVpcs")
	}); err != nil {
		return nil, err
	}

	logger.Info("success", "TotalCount", resp.TotalCount)
	if resp.TotalCount == 0 {
		return nil, nil
	}

	ret := &infrav1.VPC{}
	ret.FillFrom(&resp.Vpcs.Vpc[0])
	return ret, nil
}

func (s *VPCClient) Create(spec infrav1.VPCSpec) (string, error) {
	logger := s.WithValues("SDKAction", "Create")

	req := spec.ConvertToCreateReq()
	var resp *vpc.CreateVpcResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.CreateVpc(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "CreateVpc")
	}); err != nil {
		return "", err
	}

	logger.Info("success", "response", resp.VpcId)
	return resp.VpcId, nil
}

func (s *VPCClient) WaitReady(id string) (*infrav1.VPC, error) {
	logger := s.WithValues("SDKAction", "WaitReady", "id", id)

	var ret *infrav1.VPC
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("describing")
		var err error
		ret, err = s.Describe(id)
		if err != nil {
			logger.Info("error: " + err.Error())
			return errors.Wrap(err, "Describe")
		}

		if ret == nil {
			logger.Info("nil result")
			return retry.ErrRetry
		}

		if ret.Status != infrav1.Available {
			logger.Info("status: " + ret.Status)
			return retry.ErrRetry
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("ready")
	return ret, nil
}

func (s *VPCClient) Delete(id string) error {
	logger := s.WithValues("SDKAction", "Delete", "id", id)

	req := vpc.CreateDeleteVpcRequest()
	req.Scheme = "https"
	req.VpcId = id

	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		_, err = s.cli.DeleteVpc(req)
		if err != nil {
			logger.Info("error: " + err.Error())
			if e, ok := err.(sdkerr.Error); ok && strings.Contains(e.ErrorCode(), "Dependency") {
				return retry.ErrRetry
			}
		}
		return errors.Wrap(err, "DeleteVpc")
	}); err != nil {
		return err
	}

	logger.Info("success")
	return nil
}

func (s *VPCClient) DescribeNatGateway(id string) (*infrav1.NatGateway, error) {
	logger := s.WithValues("SDKAction", "DescribeNatGateway", "id", id)

	req := vpc.CreateDescribeNatGatewaysRequest()
	req.Scheme = "https"
	req.NatGatewayId = id

	var resp *vpc.DescribeNatGatewaysResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		resp, err = s.cli.DescribeNatGateways(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DescribeNatGateways")
	}); err != nil {
		return nil, err
	}

	logger.Info("success", "TotalCount", resp.TotalCount)
	if len(resp.NatGateways.NatGateway) == 0 {
		return nil, nil
	}

	ret := &infrav1.NatGateway{}
	ret.FillFrom(&resp.NatGateways.NatGateway[0])
	return ret, nil
}

func (s *VPCClient) CreateNatGateway(spec infrav1.NatGatewaySpec, vpcID string) (string, error) {
	logger := s.WithValues("SDKAction", "CreateNatGateway")

	req := spec.ConvertToCreateReq(vpcID)
	var resp *vpc.CreateNatGatewayResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.CreateNatGateway(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "CreateNatGateway")
	}); err != nil {
		return "", err
	}

	logger.Info("success", "NatGatewayId", resp.NatGatewayId)
	return resp.NatGatewayId, nil
}

func (s *VPCClient) WaitNatGatewayReady(id string) (*infrav1.NatGateway, error) {
	logger := s.WithValues("SDKAction", "WaitNatGatewayReady", "id", id)

	var ret *infrav1.NatGateway
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("describing")
		var err error
		ret, err = s.DescribeNatGateway(id)
		if err != nil {
			logger.Info("error: " + err.Error())
			return errors.Wrap(err, "DescribeNatGateway")
		}

		if ret == nil {
			logger.Info("nil result")
			return retry.ErrRetry
		}

		if ret.Status != infrav1.NGWAvailable {
			logger.Info("status: " + ret.Status)
			return retry.ErrRetry
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("ready")
	return ret, nil
}

func (s *VPCClient) DeleteGateway(id string) error {
	logger := s.WithValues("SDKAction", "DeleteGateway", "id", id)

	req := vpc.CreateDeleteNatGatewayRequest()
	req.Scheme = "https"
	req.NatGatewayId = id

	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		_, err = s.cli.DeleteNatGateway(req)
		if err != nil {
			if e, ok := err.(sdkerr.Error); ok && strings.Contains(e.ErrorCode(), "Dependency") {
				return retry.ErrRetry
			}
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DeleteNatGateway")
	}); err != nil {
		return err
	}

	logger.Info("success")
	return nil
}

func (s *VPCClient) DescribeEIP(id string) (*infrav1.EIP, error) {
	logger := s.WithValues("SDKAction", "DescribeEIP", "id", id)

	req := vpc.CreateDescribeEipAddressesRequest()
	req.Scheme = "https"
	req.AllocationId = id

	var resp *vpc.DescribeEipAddressesResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		resp, err = s.cli.DescribeEipAddresses(req)
		if err != nil {
			logger.Error(err, err.Error())
		}
		return errors.Wrap(err, "DescribeEipAddresses")
	}); err != nil {
		return nil, err
	}

	logger.Info("success", "TotalCount", resp.TotalCount)
	if len(resp.EipAddresses.EipAddress) == 0 {
		return nil, nil
	}

	ret := &infrav1.EIP{}
	ret.FillFrom(&resp.EipAddresses.EipAddress[0])
	return ret, nil
}

func (s *VPCClient) CreateEIP(spec infrav1.EIPSpec, vpcID string) (string, error) {
	logger := s.WithValues("SDKAction", "CreateEIP")

	req := spec.ConvertToCreateReq(vpcID)
	var resp *vpc.AllocateEipAddressResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.AllocateEipAddress(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "AllocateEipAddress")
	}); err != nil {
		return "", err
	}

	logger.Info("success", "AllocationId", resp.AllocationId)
	return resp.AllocationId, nil
}

func (s *VPCClient) WaitEIPStatus(id string, status ...string) (*infrav1.EIP, error) {
	logger := s.WithValues("SDKAction", "WaitEIPStatus", "id", id, "status", status)

	var ret *infrav1.EIP
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("describing")
		var err error
		ret, err = s.DescribeEIP(id)
		if err != nil {
			logger.Info("error: " + err.Error())
			return errors.Wrap(err, "DescribeEIP")
		}

		if ret == nil {
			logger.Info("nil result")
			return retry.ErrRetry
		}

		for _, i := range status {
			if ret.Status == i {
				return nil
			}
		}

		logger.Info(fmt.Sprintf("waiting for status: %v, now status: %v", status, ret.Status))
		return retry.ErrRetry
	}); err != nil {
		return nil, err
	}

	logger.Info("ready")
	return ret, nil
}

func (s *VPCClient) DeleteEIP(id string) error {
	logger := s.WithValues("SDKAction", "DeleteEIP", "id", id)

	req := vpc.CreateReleaseEipAddressRequest()
	req.Scheme = "https"
	req.AllocationId = id

	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		_, err = s.cli.ReleaseEipAddress(req)
		if err != nil {
			if e, ok := err.(sdkerr.Error); ok && strings.Contains(e.ErrorCode(), "Dependency") {
				return retry.ErrRetry
			}
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "ReleaseEipAddress")
	}); err != nil {
		return err
	}

	logger.Info("success")
	return nil
}

func (s *VPCClient) UnassociateEipToNatGateway(eip *infrav1.EIP, ngw *infrav1.NatGateway) error {
	logger := s.WithValues("SDKAction", "UnassociateEipToNatGateway", "eip", eip.AllocationId, "ngw", ngw.NatGatewayId)

	req := vpc.CreateUnassociateEipAddressRequest()
	req.Scheme = "https"
	req.AllocationId = eip.AllocationId
	req.InstanceId = ngw.NatGatewayId
	req.InstanceType = "Nat"

	return retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		_, err := s.cli.UnassociateEipAddress(req)
		if err != nil {
			if serr, ok := err.(sdkerr.Error); ok && serr.ErrorCode() == "InvalidIpStatus.HasBeenUsedBySnatTable" {
				return retry.ErrRetry
			}

			logger.Info("error: " + err.Error())
			return errors.Wrap(err, "UnassociateEipAddress")
		}

		logger.Info("success")
		return nil
	})
}

func (s *VPCClient) AssociateEipToNatGateway(eip *infrav1.EIP, ngw *infrav1.NatGateway) error {
	logger := s.WithValues("SDKAction", "AssociateEipToNatGateway", "eip", eip.AllocationId, "ngw", ngw.NatGatewayId)

	req := vpc.CreateAssociateEipAddressRequest()
	req.Scheme = "https"
	req.AllocationId = eip.AllocationId
	req.InstanceId = ngw.NatGatewayId
	req.InstanceType = "Nat"
	req.Mode = "NAT"

	return retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		_, err := s.cli.AssociateEipAddress(req)
		if err != nil {
			if serr, ok := err.(sdkerr.Error); ok && serr.ErrorCode() == "BIND_INSTANCE_HAVE_PORTMAP_OR_BIND_EIP" {
				return nil
			}

			logger.Info("error: " + err.Error())
			return errors.Wrap(err, "AssociateEipAddress")
		}

		logger.Info("success")
		return nil
	})
}

func (s *VPCClient) CreateSnatEntry(eip *infrav1.EIP, ngw *infrav1.NatGateway, vswID string) (string, error) {
	logger := s.WithValues("SDKAction", "CreateSnatEntry", "eip", eip.AllocationId, "ngw", ngw.NatGatewayId, "vsw", vswID)

	req := vpc.CreateCreateSnatEntryRequest()
	req.Scheme = "https"
	req.SnatTableId = ngw.SnatTableIds.SnatTableId[0]
	req.SnatIp = eip.IpAddress
	req.SourceVSwitchId = vswID
	req.SnatEntryName = ngw.Name

	var resp *vpc.CreateSnatEntryResponse
	err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		resp, err = s.cli.CreateSnatEntry(req)
		if err != nil {
			if serr, ok := err.(sdkerr.Error); ok && serr.ErrorCode() == "Forbidden.SourceVSwitchId.Duplicated" {
				return nil
			}

			logger.Info("error: " + err.Error())
			return errors.Wrap(err, "CreateSnatEntry")
		}

		logger.Info("success")
		return nil
	})
	if err != nil {
		return "", err
	}
	return resp.SnatEntryId, nil
}

func (s *VPCClient) DeleteSnatEntry(nat *infrav1.Nat) error {
	logger := s.WithValues("SDKAction", "DeleteSnatEntry", "SnatEntryId", nat.SnatEntryId)

	req := vpc.CreateDeleteSnatEntryRequest()
	req.Scheme = "https"
	req.SnatTableId = nat.NatGateway.SnatTableIds.SnatTableId[0]
	req.SnatEntryId = nat.SnatEntryId

	err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		_, err := s.cli.DeleteSnatEntry(req)
		if err != nil {
			if serr, ok := err.(sdkerr.Error); ok && serr.ErrorCode() == "Forbidden.SourceVSwitchId.Duplicated" {
				return nil
			}

			if e, ok := err.(sdkerr.Error); ok && strings.Contains(e.ErrorCode(), "NotFound") {
				return nil
			}

			logger.Info("error: " + err.Error())
			return errors.Wrap(err, "DeleteSnatEntry")
		}

		logger.Info("success")
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
