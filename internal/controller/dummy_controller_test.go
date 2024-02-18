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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	homeworkv1alpha1 "github.com/s3rj1k/dummy-controller/api/v1alpha1"
)

var _ = Describe("Dummy Controller", func() {
	Context("when reconciling a resource", func() {
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      "test-resource",
			Namespace: metav1.NamespaceDefault,
		}

		AfterEach(func() {
			resource := new(homeworkv1alpha1.Dummy)

			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("cleanup the specific resource instance Dummy")

				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		DescribeTable("should reconcile resources based on spec values",
			func(message string, expectedMessage string, invalidStatus bool) {
				By("creating a resource with a specific message in spec")

				resource := &homeworkv1alpha1.Dummy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      typeNamespacedName.Name,
						Namespace: typeNamespacedName.Namespace,
					},
					Spec: homeworkv1alpha1.DummySpec{
						Message: message,
					},
				}

				err := k8sClient.Create(ctx, resource)
				Expect(err).NotTo(HaveOccurred())

				By("reconciling the created resource")

				controllerReconciler := &DummyReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("fetching the reconciled resource to check its status")

				reconciledResource := new(homeworkv1alpha1.Dummy)

				Expect(k8sClient.Get(ctx, typeNamespacedName, reconciledResource)).To(Succeed())
				if invalidStatus {
					Expect(reconciledResource.Status.SpecEcho).NotTo(Equal(expectedMessage))
				} else {
					Expect(reconciledResource.Status.SpecEcho).To(Equal(expectedMessage))
				}

				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			},

			Entry("Valid status: strings should be equal", "foo", "foo", false),
			Entry("Valid status: strings should be equal", "bar", "bar", false),
			Entry("Valid status: empty string", "", "", false),
			Entry("Invalid status: no status when expected", "", "bar", true),
			Entry("Invalid status: status set when not expected", "bar", "", true),
		)
	})
})
