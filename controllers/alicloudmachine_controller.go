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
	rawctx "context"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "sigs.k8s.io/cluster-api-provider-alicloud/api/v1alpha2"
)

// AlicloudMachineReconciler reconciles a AlicloudMachine object
type AlicloudMachineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=alicloudmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=alicloudmachines/status,verbs=get;update;patch

func (r *AlicloudMachineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := rawctx.Background()
	machineInfra := &infrav1.AlicloudMachine{}

	err := r.Get(ctx, req.NamespacedName, machineInfra)
	if err != nil {
		return reconcile.Result{}, err
	}

	machine, err := util.GetOwnerMachine(ctx, r.Client, machineInfra.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}

	_logger := r.Log.WithValues("namespace", req.Namespace, "EcsMachineInfa", req.Name)

	if machine == nil {
		_logger.Info("Machine Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	clusterInfra := &infrav1.AlicloudCluster{}
	cluster := &clusterv1.Cluster{}
	//if machine.DeletionTimestamp == nil {
	cluster, err = util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		_logger.Info("cluster not found")
		return reconcile.Result{}, nil
	}

	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}, clusterInfra); err != nil {
		_logger.Info("ClusterInfra is not available")
		return reconcile.Result{}, nil
	}
	//}

	processer := &MachineProcesser{
		AlicloudMachineReconciler: *r,
		cluster:                   cluster,
		machine:                   machine,
		clusterInfra:              clusterInfra,
		machineInfra:              machineInfra,
		result:                    ctrl.Result{},
	}
	processer.Log = _logger

	processer.init()

	processer.sync()
	return processer.result, processer.err

}

func (r *AlicloudMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.AlicloudMachine{}).
		Watches(
			&source.Kind{Type: &clusterv1.Machine{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("AlicloudMachine")),
			},
		).
		Complete(r)
}
