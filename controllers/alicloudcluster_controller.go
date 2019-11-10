/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
)

// AlicloudClusterReconciler reconciles a AlicloudCluster object
type AlicloudClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=alicloudclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=alicloudclusters/status,verbs=get;update;patch

func (r *AlicloudClusterReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	logger := r.Log.WithValues("AlicloudClusterReconciler", req.NamespacedName)

	alicloudCluster := &infrav1.AlicloudCluster{}
	if err := r.Get(ctx, req.NamespacedName, alicloudCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	cluster, err := util.GetOwnerCluster(ctx, r.Client, alicloudCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "GetOwnerCluster")
	}
	if cluster == nil {
		logger.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}

	processor, err := NewClusterProcessor(logger, alicloudCluster.Spec.RegionId, r.Client, cluster, alicloudCluster)
	if err != nil {
		logger.Error(err, "NewClusterProcessor error")
		return reconcile.Result{}, errors.Wrap(err, "NewClusterProcessor")
	}

	defer func() {
		if err := processor.Close(); err != nil && reterr == nil {
			logger.Error(err, "processor.Close error")
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !alicloudCluster.DeletionTimestamp.IsZero() {
		if ret, err := processor.ReconcileDelete(); err != nil {
			logger.Error(err, "ReconcileDelete error")
			return ret, errors.Wrap(err, "ReconcileDelete")
		}
		return ctrl.Result{}, nil
	}

	// Handle non-deleted clusters
	if ret, err := processor.ReconcileNormal(); err != nil {
		alicloudCluster.Status.Reason = err.Error()
		alicloudCluster.Status.Ready = false

		logger.Error(err, "ReconcileNormal error")
		return ret, errors.Wrap(err, "ReconcileNormal")
	}

	alicloudCluster.Status.Message = "success"
	alicloudCluster.Status.Reason = ""
	alicloudCluster.Status.Ready = true

	return ctrl.Result{}, nil
}

func (r *AlicloudClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.AlicloudCluster{}).
		Complete(r)
}
