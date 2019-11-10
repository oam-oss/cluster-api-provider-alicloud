package aliyun

import (
	"fmt"
	"strings"

	sdkerr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg/aliyun/retry"
)

func NewSLBClient(logger logr.Logger, regionID string) (*SLBClient, error) {
	cli, err := slb.NewClientWithAccessKey(regionID, AccessKeyId, AccessKeySecret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create slb client")
	}
	return &SLBClient{
		Logger: logger.WithValues("client", "slb"),
		cli:    cli,
	}, nil
}

type SLBClient struct {
	logr.Logger
	cli *slb.Client
}

func (s *SLBClient) Describe(id string) (*infrav1.SLB, error) {
	logger := s.WithValues("SDKAction", "Describe", "id", id)

	req := slb.CreateDescribeLoadBalancersRequest()
	req.Scheme = "https"
	req.LoadBalancerId = id

	var resp *slb.DescribeLoadBalancersResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		resp, err = s.cli.DescribeLoadBalancers(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DescribeLoadBalancers")
	}); err != nil {
		return nil, err
	}

	logger.Info("success", "TotalCount", resp.TotalCount)
	if resp.TotalCount == 0 {
		return nil, nil
	}

	ret := &infrav1.SLB{}
	ret.FillFrom(&resp.LoadBalancers.LoadBalancer[0])
	return ret, nil
}

func (s *SLBClient) Create(spec infrav1.SLBSpec, vpcId string) (string, error) {
	logger := s.WithValues("SDKAction", "Create")

	req := spec.ConvertToCreateReq(vpcId)
	var resp *slb.CreateLoadBalancerResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.CreateLoadBalancer(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "CreateLoadBalancer")
	}); err != nil {
		return "", err
	}

	logger.Info("success", "LoadBalancerId", resp.LoadBalancerId)
	return resp.LoadBalancerId, nil
}

func (s *SLBClient) WaitReady(id string) (*infrav1.SLB, error) {
	logger := s.WithValues("SDKAction", "WaitReady", "id", id)

	var ret *infrav1.SLB
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

		if ret.LoadBalancerStatus != infrav1.SLBActive {
			logger.Info("status: " + ret.LoadBalancerStatus)
			return retry.ErrRetry
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("ready")
	return ret, nil
}

func (s *SLBClient) Delete(id string) error {
	logger := s.WithValues("SDKAction", "Delete")

	req := slb.CreateDeleteLoadBalancerRequest()
	req.Scheme = "https"
	req.LoadBalancerId = id

	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		_, err = s.cli.DeleteLoadBalancer(req)
		if err != nil {
			if e, ok := err.(sdkerr.Error); ok && strings.Contains(e.ErrorCode(), "Dependency") {
				return retry.ErrRetry
			}
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DeleteLoadBalancer")
	}); err != nil {
		return err
	}

	logger.Info("success")
	return nil
}

func (s *SLBClient) DescribeServerGroup(slbID string) ([]string, error) {
	logger := s.WithValues("SDKAction", "DescribeServerGroup")

	req := slb.CreateDescribeVServerGroupsRequest()
	req.LoadBalancerId = slbID
	var vgResp *slb.DescribeVServerGroupsResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		vgResp, err = s.cli.DescribeVServerGroups(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DescribeVServerGroups")
	}); err != nil {
		return nil, err
	}

	var list []string
	for _, v := range vgResp.VServerGroups.VServerGroup {
		list = append(list, v.VServerGroupId)
	}

	return list, nil
}

func (s *SLBClient) CreateServerGroup(spec infrav1.SLBSpec, slbID string) (string, error) {
	logger := s.WithValues("SDKAction", "CreateServerGroup")

	req := spec.ConvertToCreateSLBVGReq(slbID)
	var resp *slb.CreateVServerGroupResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.CreateVServerGroup(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "CreateVServerGroup")
	}); err != nil {
		return "", err
	}

	logger.Info("success", "VServerGroupId", resp.VServerGroupId)
	return resp.VServerGroupId, nil
}

func (s *SLBClient) CreateTCPListener(spec infrav1.SLBSpec, slbID string, vgID string) error {
	logger := s.WithValues("SDKAction", "CreateTCPListener")

	req := spec.ConvertToCreateSLBTCPListenerReq(slbID, vgID)
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		_, err = s.cli.CreateLoadBalancerTCPListener(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "CreateLoadBalancerTCPListener")
	}); err != nil {
		return err
	}

	logger.Info("success")
	return nil
}

func (s *SLBClient) StartListener(slbID string) error {
	logger := s.WithValues("SDKAction", "StartListener")

	req := slb.CreateDescribeLoadBalancerTCPListenerAttributeRequest()
	req.Scheme = "https"
	req.LoadBalancerId = slbID
	req.ListenerPort = requests.NewInteger(6443)

	resp, err := s.cli.DescribeLoadBalancerTCPListenerAttribute(req)
	if err != nil {
		return err
	}

	if resp.Status != "starting" && resp.Status != "running" {
		startReq := slb.CreateStartLoadBalancerListenerRequest()
		startReq.Scheme = "https"

		startReq.LoadBalancerId = slbID
		startReq.ListenerPort = requests.NewInteger(6443)

		if err := retry.Try(retry.DefaultBackOf, func() error {
			logger.Info("requesting", "request", startReq)
			_, err := s.cli.StartLoadBalancerListener(startReq)
			if err != nil {
				logger.Info("error: " + err.Error())
			}
			return errors.Wrap(err, "StartLoadBalancerListener")
		}); err != nil {
			return err
		}

	}

	return nil
}

func (s *SLBClient) VGAddBackendServers(vgID, instanceID, port, desc string) (*slb.AddVServerGroupBackendServersResponse, error) {
	logger := s.WithValues("SDKAction", "VGAddBackendServers")

	req := slb.CreateAddVServerGroupBackendServersRequest()
	req.Scheme = "https"
	req.VServerGroupId = vgID
	req.BackendServers = fmt.Sprintf(`[{ "ServerId": "%v", "Port": "%v", "Weight": "100", "Type": "ecs", "Description":"%v" }]`, instanceID, port, desc)

	var resp *slb.AddVServerGroupBackendServersResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.AddVServerGroupBackendServers(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "AddVServerGroupBackendServers")
	}); err != nil {
		return nil, err
	}

	logger.Info("success", "BackendServers", resp.BackendServers)

	return resp, nil
}
