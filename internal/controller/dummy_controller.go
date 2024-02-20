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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	homeworkv1alpha1 "github.com/s3rj1k/dummy-controller/api/v1alpha1"
)

const (
	TimeToRequeueOnSuccess = 5 * time.Minute

	PodLabelKey = "homework.interview.me/pod"

	PodContainerName  = "dummy-object-bound-container"
	PodContainerImage = "nginx:latest"
)

// DummyControllerConfig contains additional config options for controller
type DummyControllerConfig struct {
	TimeToRequeueOnSuccess time.Duration

	PodContainerName  string
	PodContainerImage string
}

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config DummyControllerConfig
}

func (r *DummyReconciler) getBoundPodName(o metav1.ObjectMeta) string {
	return fmt.Sprintf("%s-pod-%s", r.Config.PodContainerName, o.Name)
}

func (r *DummyReconciler) updateStatus(ctx context.Context, obj *homeworkv1alpha1.Dummy) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	pod := new(corev1.Pod)

	err := r.Get(ctx, types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      r.getBoundPodName(obj.ObjectMeta),
	}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.V(4).Info("Corresponding Pod object not found.")

			return ctrl.Result{Requeue: true}, nil
		}

		reqLogger.Error(err, "Failed to get corresponding Pod object")

		return ctrl.Result{Requeue: true}, err
	}

	if obj.Status.SpecEcho == obj.Spec.Message && obj.Status.PodStatus == pod.Status.Phase {
		return ctrl.Result{RequeueAfter: r.Config.TimeToRequeueOnSuccess}, nil
	}

	obj.Status.SpecEcho = obj.Spec.Message
	obj.Status.PodStatus = pod.Status.Phase

	err = r.Status().Update(ctx, obj)
	if err != nil {
		reqLogger.Error(err, "Failed to update Dummy object status")

		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{RequeueAfter: r.Config.TimeToRequeueOnSuccess}, nil
}

//+kubebuilder:rbac:groups=homework.interview.me,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homework.interview.me,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homework.interview.me,resources=dummies/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	reqLogger.V(4).Info("Reconciling Dummy object")

	dummy := new(homeworkv1alpha1.Dummy)

	err := r.Get(ctx, req.NamespacedName, dummy)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.V(4).Info("Dummy object not found.")

			return ctrl.Result{}, nil
		}

		reqLogger.Error(err, "Failed to get Dummy object")

		return ctrl.Result{}, err
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.getBoundPodName(dummy.ObjectMeta),
			Namespace: req.Namespace,
			Labels: map[string]string{
				PodLabelKey: req.Name,
			},
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, pod, func() error {
		if pod.CreationTimestamp.IsZero() {
			pod.Spec.Containers = []corev1.Container{
				{
					Name:  r.Config.PodContainerName,
					Image: r.Config.PodContainerImage,
				},
			}
		} else {
			for i, container := range pod.Spec.Containers {
				if container.Name != r.Config.PodContainerName {
					continue
				}

				if container.Image != r.Config.PodContainerImage {
					pod.Spec.Containers[i].Image = r.Config.PodContainerImage
				}

				break
			}
		}

		err := controllerutil.SetControllerReference(dummy, pod, r.Scheme)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		reqLogger.Error(err, "Failed to ensure corresponding Pod object")

		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("Dummy object", "Message", dummy.Spec.Message)

	return r.updateStatus(ctx, dummy)
}

func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homeworkv1alpha1.Dummy{}).
		Watches(&corev1.Pod{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&homeworkv1alpha1.Dummy{},
			),
		).
		Complete(r)
}
