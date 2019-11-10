package retry

import (
	"time"

	sdkerr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

var DefaultBackOf = wait.Backoff{
	Duration: time.Second,
	Factor:   2,
	Steps:    32,
	Jitter:   4,
	Cap:      20 * time.Second,
}

var ErrRetry = errors.New("retry")

func Try(backoff wait.Backoff, fn func() error) error {
	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		if err == nil {
			return true, nil
		}

		if errors.Cause(err) == ErrRetry {
			return false, nil
		}

		if err, ok := err.(sdkerr.Error); ok {
			// timeout or server errors should retry
			if err.ErrorCode() == sdkerr.TimeoutErrorCode || err.HttpStatus() >= 500 {
				return false, nil
			}
			// ignore not found error
			if err.HttpStatus() == 404 {
				return true, nil
			}
		}

		// errors can't retry
		return false, err
	})
}
