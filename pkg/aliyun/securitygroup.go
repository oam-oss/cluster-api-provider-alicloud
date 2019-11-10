package aliyun

import (
	"strings"

	sdkerr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg/aliyun/retry"
)

func NewSecurityGroupClient(logger logr.Logger, regionID string) (*SecurityGroupClient, error) {
	cli, err := ecs.NewClientWithAccessKey(regionID, AccessKeyId, AccessKeySecret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create securityGroup client")
	}
	return &SecurityGroupClient{
		Logger: logger.WithValues("client", "securityGroup"),
		cli:    cli,
	}, nil
}

type SecurityGroupClient struct {
	logr.Logger
	cli *ecs.Client
}

func (s *SecurityGroupClient) Describe(id string) (*infrav1.SecurityGroup, error) {
	logger := s.WithValues("SDKAction", "Describe", "id", id)

	req := ecs.CreateDescribeSecurityGroupsRequest()
	req.Scheme = "https"
	req.SecurityGroupId = id

	var resp *ecs.DescribeSecurityGroupsResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		resp, err = s.cli.DescribeSecurityGroups(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DescribeSecurityGroups")
	}); err != nil {
		return nil, err
	}

	logger.Info("success", "response", resp, "TotalCount", resp.TotalCount)
	if resp.TotalCount == 0 {
		return nil, nil
	}

	ret := &infrav1.SecurityGroup{}
	ret.FillFrom(&resp.SecurityGroups.SecurityGroup[0])
	return ret, nil
}

func (s *SecurityGroupClient) Create(spec infrav1.SecurityGroupSpec, vpcId string) (string, error) {
	logger := s.WithValues("SDKAction", "Create")

	req := spec.ConvertToCreateReq(vpcId)
	var resp *ecs.CreateSecurityGroupResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.CreateSecurityGroup(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "CreateSecurityGroup")
	}); err != nil {
		return "", err
	}

	logger.Info("success", "SecurityGroupId", resp.SecurityGroupId)

	rules := spec.ConvertToAuthorizeSecurityGroupRequest(resp.SecurityGroupId)
	for _, rule := range rules {
		logger := s.WithValues("SDKAction", "CreateRule")
		if err := retry.Try(retry.DefaultBackOf, func() error {
			logger.Info("requesting", "request", req)
			_, err := s.cli.AuthorizeSecurityGroup(rule)
			if err != nil {
				logger.Info("error: " + err.Error())
			}
			return errors.Wrap(err, "AuthorizeSecurityGroup")
		}); err != nil {
			return "", err
		}
	}

	return resp.SecurityGroupId, nil
}

func (s *SecurityGroupClient) WaitReady(id string) (*infrav1.SecurityGroup, error) {
	logger := s.WithValues("SDKAction", "WaitReady", "id", id)

	var ret *infrav1.SecurityGroup
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

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("ready")
	return ret, nil
}

func (s *SecurityGroupClient) Delete(id string) error {
	logger := s.WithValues("SDKAction", "Delete")

	req := ecs.CreateDeleteSecurityGroupRequest()
	req.Scheme = "https"
	req.SecurityGroupId = id

	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		_, err = s.cli.DeleteSecurityGroup(req)
		if err != nil {
			if e, ok := err.(sdkerr.Error); ok && strings.Contains(e.ErrorCode(), "Dependency") {
				return retry.ErrRetry
			}
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DeleteSecurityGroup")
	}); err != nil {
		return err
	}

	logger.Info("success")
	return nil
}
