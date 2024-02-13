/*
Copyright 2023 freedisch.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netwv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	// "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/freedisch/k8s-operator/api/v1alpha1"
)

// ExposeReconciler reconciles a Expose object
type ExposeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.core.k8s-operator.io,resources=exposes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.core.k8s-operator.io,resources=exposes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.core.k8s-operator.io,resources=exposes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Expose object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ExposeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	expose := &apiv1alpha1.Expose{}
	err := r.Get(ctx, req.NamespacedName, expose)
	if err != nil {
		return ctrl.Result{}, nil
	}

	deploymentName := expose.Spec.Deployment[0].Name
	replicas := expose.Spec.Deployment[0].Replicas
	component := expose.Spec.Deployment[0].Component
	containers := expose.Spec.Deployment[0].Containers
	serviceName := expose.Spec.Service[0].Name
	port := expose.Spec.Service[0].Port
	ingressName := expose.Spec.Ingress[0].Name
	path := expose.Spec.Ingress[0].Path

	finalizerName := "api.core.k8s-operator.io/finalizer"

	if expose.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(expose, finalizerName) {
			controllerutil.AddFinalizer(expose, finalizerName)
			if err := r.Update(ctx, expose); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(expose, finalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteResources(ctx, expose, deploymentName, serviceName, ingressName); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(expose, finalizerName)
			if err := r.Update(ctx, expose); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Print deployment and container information
	fmt.Printf("Deployment: %s, Replicas: %d \n", deploymentName, replicas)
	for _, container := range containers {
		fmt.Printf("Container: %s, Image: %s \n", container.Name, container.Image)
	}

	// Create deployment
	existingDeployment := &appsv1.Deployment{}
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Namespace: expose.Namespace, Name: deploymentName}, existingDeployment)
	if err != nil {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: expose.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(replicas),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": component,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": component,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  containers[0].Name,
								Image: containers[0].Image,
							},
						},
					},
				},
			},
		}
		fmt.Println("Creating deployment...")
		err = r.Create(ctx, deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("Created deployment %s \n", deployment.Name)
	} else {
		fmt.Printf("Deployment %s already exists \n", deploymentName)
	}

	// Create Service
	existingService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Namespace: expose.Namespace, Name: serviceName}, existingService)
	if err != nil {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: expose.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: getDeploymentLabels(*deployment),
				Ports: []corev1.ServicePort{
					{
						Name: "http",
						Port: port,
					},
				},
			},
		}
		fmt.Println("Creating service...")
		err = r.Create(ctx, service)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("Created service %s \n", service.Name)
	} else {
		fmt.Printf("Service %s already exists \n", serviceName)
	}

	// Create Ingress
	pathType := "Prefix"
	existingIngress := &netwv1.Ingress{}
	err = r.Get(ctx, client.ObjectKey{Namespace: expose.Namespace, Name: ingressName}, existingIngress)
	if err != nil {
		ingress := &netwv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ingressName,
				Namespace: expose.Namespace,
				Annotations: map[string]string{
					"nginx.ingress.kubernetes.io/app-root": "/",
				},
			},
			Spec: netwv1.IngressSpec{
				Rules: []netwv1.IngressRule{
					netwv1.IngressRule{
						IngressRuleValue: netwv1.IngressRuleValue{
							HTTP: &netwv1.HTTPIngressRuleValue{
								Paths: []netwv1.HTTPIngressPath{
									netwv1.HTTPIngressPath{
										Path:     path,
										PathType: (*netwv1.PathType)(&pathType),
										Backend: netwv1.IngressBackend{
											Service: &netwv1.IngressServiceBackend{
												Name: serviceName,
												Port: netwv1.ServiceBackendPort{
													Number: port,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		fmt.Println("Creating ingress...")
		err = r.Create(ctx, ingress)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("Created ingress %s \n", ingress.Name)
	} else {
		fmt.Printf("Ingress %s already exists \n", ingressName)
	}

	return ctrl.Result{}, nil

	// Periodic reconcilation
	// return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExposeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Expose{}).
		Complete(r)
}

func int32Ptr(i int32) *int32 {
	return &i
}

func getDeploymentLabels(deployment appsv1.Deployment) map[string]string {
	return deployment.Spec.Template.Labels
}

func (r *ExposeReconciler) deleteResources(ctx context.Context, expose *apiv1alpha1.Expose, deploymentName string, serviceName string, ingressName string) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: expose.Namespace,
		},
	}
	fmt.Println("Deleting deployment...")
	if err := r.Delete(ctx, deployment); err != nil {
		fmt.Printf("Error in deleting a deployment %s \n", deployment.Name)
		return nil
	}
	fmt.Printf("Deleted deployment %s \n", deployment.Name)

	// Delete Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: expose.Namespace,
		},
	}
	fmt.Println("Deleting service...")
	if err := r.Delete(ctx, service); err != nil {
		fmt.Printf("Error in deleting a service %s \n", service.Name)
		return nil
	}
	fmt.Printf("Deleted service %s \n", service.Name)

	// Delete Ingress
	ingress := &netwv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: expose.Namespace,
		},
	}
	fmt.Println("Deleting ingress...")
	if err := r.Delete(ctx, ingress); err != nil {
		fmt.Printf("Error in deleting a ingress %s \n", ingress.Name)
		return nil
	}
	fmt.Printf("Deleted ingress %s \n", ingress.Name)

	return nil
}
