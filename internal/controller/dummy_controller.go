/*
Copyright 2024 s3rj1k.

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

package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	homeworkv1alpha1 "github.com/s3rj1k/dummy-controller/api/v1alpha1"
)

const TimeToRequeueOnSuccess = 5 * time.Minute

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homework.interview.me,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homework.interview.me,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homework.interview.me,resources=dummies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	reqLogger.V(6).Info("Reconciling Dummy object")

	dummy := new(homeworkv1alpha1.Dummy)

	err := r.Get(ctx, req.NamespacedName, dummy)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.V(8).Info("Dummy object not found.")

			return ctrl.Result{}, nil
		}

		reqLogger.Error(err, "Failed to get Dummy object")

		return ctrl.Result{}, err
	}

	reqLogger.Info("Dummy object", "Message", dummy.Spec.Message)

	if dummy.Status.SpecEcho == dummy.Spec.Message {
		return ctrl.Result{RequeueAfter: TimeToRequeueOnSuccess}, nil
	}

	dummy.Status.SpecEcho = dummy.Spec.Message

	err = r.Status().Update(ctx, dummy)
	if err != nil {
		reqLogger.Error(err, "Failed to update Dummy object status")

		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: TimeToRequeueOnSuccess}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homeworkv1alpha1.Dummy{}).
		Complete(r)
}
