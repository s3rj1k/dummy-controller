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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	homeworkv1alpha1 "github.com/s3rj1k/dummy-controller/api/v1alpha1"
)

var _ = Describe("Dummy Controller", Ordered, func() {
	ctx := context.Background()

	resourceTypeNamespacedName := types.NamespacedName{
		Name:      "test-dummy",
		Namespace: metav1.NamespaceDefault,
	}

	BeforeEach(func() {
		// Setup code if needed
	})

	DescribeTable("Should reconcile resources based on spec values",
		func(message string, expectedMessage string, invalidStatus bool) {
			By("Creating a resource with a specific message in spec")
			resource := &homeworkv1alpha1.Dummy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceTypeNamespacedName.Name,
					Namespace: resourceTypeNamespacedName.Namespace,
				},
				Spec: homeworkv1alpha1.DummySpec{
					Message: message,
				},
			}
			err := k8sClient.Create(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the created resource")
			r := &DummyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Config: DummyControllerConfig{
					PodContainerName:  PodContainerName,
					PodContainerImage: "google/pause:latest",
				},
			}
			_, err = r.Reconcile(ctx, reconcile.Request{
				NamespacedName: resourceTypeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Fetching the reconciled resource to check its status")
			reconciledResource := new(homeworkv1alpha1.Dummy)
			Expect(k8sClient.Get(ctx, resourceTypeNamespacedName, reconciledResource)).To(Succeed())
			if invalidStatus {
				Expect(reconciledResource.Status.SpecEcho).NotTo(Equal(expectedMessage))
			} else {
				Expect(reconciledResource.Status.SpecEcho).To(Equal(expectedMessage))
			}

			By("Checking for the presence of a Pod after reconcile")
			pod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      r.getBoundPodName(resource.ObjectMeta),
				Namespace: metav1.NamespaceDefault,
			}, pod)
			Expect(err).NotTo(HaveOccurred(), "Pod should exist after reconcile")
		},

		Entry("Valid status: strings should be equal", "foo", "foo", false),
		Entry("Valid status: strings should be equal", "bar", "bar", false),
		Entry("Valid status: empty string", "", "", false),
		Entry("Invalid status: no status when expected", "", "bar", true),
		Entry("Invalid status: status set when not expected", "bar", "", true),
	)

	AfterEach(func() {
		resource := new(homeworkv1alpha1.Dummy)

		err := k8sClient.Get(ctx, resourceTypeNamespacedName, resource)
		if err != nil {
			if errors.IsNotFound(err) {
				return
			} else {
				Expect(err).NotTo(HaveOccurred(), "Unexpected error when retrieving the resource")
			}
		}

		By("Cleanup the specific resource instance Dummy")
		Expect(k8sClient.Delete(ctx, resource, []client.DeleteOption{
			client.PropagationPolicy(metav1.DeletePropagationBackground),
		}...)).To(Succeed())

		By("Waiting for Dummy resource to be deleted")
		Eventually(func() bool {
			pod := new(corev1.Pod)

			err := k8sClient.Get(ctx, resourceTypeNamespacedName, pod)
			if errors.IsNotFound(err) {
				return true
			} else if err != nil {
				return false
			} else {
				if pod.ObjectMeta.DeletionTimestamp != nil {
					return true
				}
			}
			return false
		}, "1m", "10s").Should(BeTrue(), "Dummy resource should be deleted or in terminating state within the timeout period")
	})
})
