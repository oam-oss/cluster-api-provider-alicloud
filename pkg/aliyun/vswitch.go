package aliyun

import (
	"strings"

	sdkerr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
	"sigs.k8s.io/cluster-api-provider-alicloud/pkg/aliyun/retry"
)

func NewVSwitchClient(logger logr.Logger, regionID string) (*VSwitchClient, error) {
	cli, err := vpc.NewClientWithAccessKey(regionID, AccessKeyId, AccessKeySecret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vswitch client")
	}
	return &VSwitchClient{
		Logger: logger.WithValues("client", "vswitch"),
		cli:    cli,
	}, nil
}

type VSwitchClient struct {
	logr.Logger
	cli *vpc.Client
}

func (s *VSwitchClient) Describe(id string) (*infrav1.VSwitch, error) {
	logger := s.WithValues("SDKAction", "Describe", "id", id)
	req := vpc.CreateDescribeVSwitchesRequest()
	req.Scheme = "https"
	req.VSwitchId = id

	var resp *vpc.DescribeVSwitchesResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		resp, err = s.cli.DescribeVSwitches(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DescribeVSwitches")
	}); err != nil {
		return nil, err
	}

	logger.Info("success", "TotalCount", resp.TotalCount)
	if resp.TotalCount == 0 {
		return nil, nil
	}

	ret := &infrav1.VSwitch{}
	ret.FillFrom(&resp.VSwitches.VSwitch[0])
	return ret, nil
}

func (s *VSwitchClient) Create(spec infrav1.VSwitchSpec, zoneID string, vpcId string) (string, error) {
	logger := s.WithValues("SDKAction", "Create")

	req := spec.ConvertToCreateReq(vpcId, zoneID)
	var resp *vpc.CreateVSwitchResponse
	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting", "request", req)
		var err error
		resp, err = s.cli.CreateVSwitch(req)
		if err != nil {
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "CreateVSwitch")
	}); err != nil {
		return "", err
	}

	logger.Info("success", "response", resp.VSwitchId)
	return resp.VSwitchId, nil
}

func (s *VSwitchClient) WaitReady(id string) (*infrav1.VSwitch, error) {
	logger := s.WithValues("SDKAction", "WaitReady", "id", id)

	var ret *infrav1.VSwitch
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

func (s *VSwitchClient) Delete(id string) error {
	logger := s.WithValues("SDKAction", "Delete", "id", id)

	req := vpc.CreateDeleteVSwitchRequest()
	req.Scheme = "https"
	req.VSwitchId = id

	if err := retry.Try(retry.DefaultBackOf, func() error {
		logger.Info("requesting")
		var err error
		_, err = s.cli.DeleteVSwitch(req)
		if err != nil {
			if e, ok := err.(sdkerr.Error); ok {
				if strings.Contains(e.Message(), "not found") {
					return nil
				}
				if strings.Contains(e.ErrorCode(), "Dependency") {
					return retry.ErrRetry
				}
			}
			logger.Info("error: " + err.Error())
		}
		return errors.Wrap(err, "DeleteVSwitch")
	}); err != nil {
		return err
	}

	s.Info("success")
	return nil
}
