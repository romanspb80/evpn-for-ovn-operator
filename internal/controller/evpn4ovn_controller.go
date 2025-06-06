/*
Copyright 2025.

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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	evpnapiv1alpha1 "github.com/romanspb80/evpn-for-ovn-operator/api/v1alpha1"
)

const apiPortName = "http"

const imageOvsdbAgent = "28041980/ovsdb-agent:latest"
const imageMpbgpAgent = "28041980/mpbgp-agent:latest"
const imageEvpnApi = "28041980/evpn-api:latest"
const finalizer = "domain-x.com/evpn4ovn_controller_finalizer"

// Evpn4OvnReconciler reconciles a Evpn4Ovn object
type Evpn4OvnReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=evpn-api.domain-x.com,resources=evpn4ovns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=evpn-api.domain-x.com,resources=evpn4ovns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=evpn-api.domain-x.com,resources=evpn4ovns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Evpn4Ovn object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *Evpn4OvnReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	var evpn4ovn evpnapiv1alpha1.Evpn4Ovn

	if err := r.Get(ctx, req.NamespacedName, &evpn4ovn); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		l.Error(err, "unable to fetch Application")
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(&evpn4ovn, finalizer) {
		l.Info("Adding Finalizer")
		controllerutil.AddFinalizer(&evpn4ovn, finalizer)
		return ctrl.Result{}, r.Update(ctx, &evpn4ovn)
	}

	//if !evpn4ovn.DeletionTimestamp.IsZero() {
	//	l.Info("Evpn4Ovn is being deleted")
	//	return r.reconcileDelete(ctx, &evpn4ovn)
	//}
	l.Info("Evpn4Ovn is being created")
	return r.reconcileCreate(ctx, &evpn4ovn)
}

func (r *Evpn4OvnReconciler) reconcileCreate(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Creating ApiReplicaSe")
	err := r.createOrUpdateApiReplicaSet(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating OvsdbAgentReplicaSet")
	err = r.createOrUpdateOvsdbAgentReplicaSet(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating MpbgpAgentDaemonset")
	err = r.createOrUpdateMpbgpAgentDaemonset(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating API service")
	err = r.createApiService(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating RabbitMQ service")
	err = r.createRabbitmqService(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating RabbitMQ endpoint")
	err = r.createRabbitmqEndpoint(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating BGP endpoint")
	err = r.createBgpEndpoint(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating OVSDB service")
	err = r.createRabbitmqService(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating OVSDB endpoint")
	err = r.createRabbitmqEndpoint(ctx, evpn4ovn)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Evpn4OvnReconciler) createOrUpdateApiReplicaSet(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	var apiReplicaSet appsv1.ReplicaSet
	// ToDo: set replicas as a parrameter
	replicas := int32(1)
	apiReplicaSet = appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "evpn-api",
			//Namespace: evpn4ovn.ObjectMeta.Name,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "evpn-api",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "evpn-api",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "evpn-api",
							Image: imageEvpnApi,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "app-settings",
									MountPath: "/config",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          apiPortName,
									ContainerPort: evpn4ovn.Spec.Api.Port,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{Name: "app-settings", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "app-settings",
							},
						}}},
					},
				},
			},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &apiReplicaSet, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create replicaset of API: %v", err)
	}

	return nil

}

func (r *Evpn4OvnReconciler) createOrUpdateOvsdbAgentReplicaSet(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	var ovsdbAgentReplicaSet appsv1.ReplicaSet
	replicas := int32(1)
	rsName := types.NamespacedName{Name: "ovsdb-agent"} //Namespace: evpn4ovn.ObjectMeta.Name,

	if err := r.Get(ctx, rsName, &ovsdbAgentReplicaSet); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("unable to fetch ReplicaSet: %v", err)
		}
		/*If there is no Deployment, we will create it.
		An essential section in the definition is OwnerReferences, as it indicates to k8s that
		an Application is creating this resource.
		This is how k8s knows that when we remove an Application, it must also remove
		all the resources it created.
		Another important detail is that we use data from our Application to create the Deployment,
		such as image information, port, and replicas.
		*/
		if apierrors.IsNotFound(err) {
			// api replicaSet

		}
		err = r.Create(ctx, &ovsdbAgentReplicaSet)
		if err != nil {
			return fmt.Errorf("unable to create replicaset of OVSDB Agent: %v", err)
		}

		ovsdbAgentReplicaSet = appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ovsdb-agent",
				Namespace: evpn4ovn.ObjectMeta.Name,
				Labels:    map[string]string{"app": "ovsdb-agent"},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: evpn4ovn.APIVersion,
						Kind:       evpn4ovn.Kind,
						Name:       evpn4ovn.Name,
						UID:        evpn4ovn.UID,
					},
				},
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "ovsdb-agent"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "ovsdb-agent"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  evpn4ovn.ObjectMeta.Name + "-container",
								Image: imageOvsdbAgent,
								Env: []corev1.EnvVar{
									{
										Name:  "OVN_PROVIDER",
										Value: "microovn",
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "app-settings",
										MountPath: "/config",
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "app-settings",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "app-settings",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		err = r.Create(ctx, &ovsdbAgentReplicaSet)
		if err != nil {
			return fmt.Errorf("unable to create replicaset of OVSDB Agent: %v", err)
		}

	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &ovsdbAgentReplicaSet, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create replicaset of API: %v", err)
	}

	return nil
}

func (r *Evpn4OvnReconciler) createOrUpdateMpbgpAgentDaemonset(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	mpbgpAgentDaemonset := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      evpn4ovn.ObjectMeta.Name + "-daemonset",
			Namespace: evpn4ovn.ObjectMeta.Name,
			Labels:    map[string]string{"app": "mpbgp-agent"},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "mpbgp-agent"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "mpbgp-agent"},
				},
				Spec: corev1.PodSpec{
					HostNetwork: true,
					Containers: []corev1.Container{
						{
							Name:  evpn4ovn.ObjectMeta.Name + "-container",
							Image: imageMpbgpAgent,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "app-settings", MountPath: "/config"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{Name: "app-settings", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "app-settings",
							},
						}}},
					},
				},
			},
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &mpbgpAgentDaemonset, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create Daemonset of MPBGP-Agent: %v", err)
	}

	return nil

}

func (r *Evpn4OvnReconciler) createApiService(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	srv := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "api-svc",
			//Namespace: evpn4ovn.ObjectMeta.Name,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "evpn-api"},
			Ports: []corev1.ServicePort{
				{
					Port:       evpn4ovn.Spec.Api.Port,
					TargetPort: intstr.FromString("http"),
				},
			},
		},
		Status: corev1.ServiceStatus{},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &srv, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create API Service: %v", err)
	}
	return nil
}

func (r *Evpn4OvnReconciler) createRabbitmqService(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	srv := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: evpn4ovn.Spec.RabbitmqService.Name,
			//Namespace: evpn4ovn.ObjectMeta.Name,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			//ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			Ports: []corev1.ServicePort{
				{
					Name:     evpn4ovn.Spec.RabbitmqService.Endpoints.Ports[0].Name,
					Port:     evpn4ovn.Spec.RabbitmqService.Endpoints.Ports[0].Port,
					Protocol: corev1.ProtocolTCP,
					NodePort: 0,
				},
			},
		},
		Status: corev1.ServiceStatus{},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &srv, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create RabbitMQ Service: %v", err)
	}
	return nil
}

func (r *Evpn4OvnReconciler) createRabbitmqEndpoint(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	ep := corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: evpn4ovn.Spec.RabbitmqService.Name,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: evpn4ovn.Spec.RabbitmqService.Endpoints.IP[0]},
				},
				Ports: []corev1.EndpointPort{
					{
						Name: evpn4ovn.Spec.RabbitmqService.Endpoints.Ports[0].Name,
						Port: evpn4ovn.Spec.RabbitmqService.Endpoints.Ports[0].Port,
					},
				},
			},
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &ep, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create RabbitMQ endpoint: %v", err)
	}
	return nil
}

func (r *Evpn4OvnReconciler) createBgpEndpoint(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	ep := corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: evpn4ovn.Spec.BgpService.Name,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: evpn4ovn.Spec.BgpService.Endpoints.IP[0]},
				},
				Ports: []corev1.EndpointPort{
					{
						Name: evpn4ovn.Spec.BgpService.Endpoints.Ports[0].Name,
						Port: evpn4ovn.Spec.BgpService.Endpoints.Ports[0].Port,
					},
				},
			},
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &ep, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create BGP endpoint: %v", err)
	}
	return nil
}

func (r *Evpn4OvnReconciler) createOvsdbService(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	ports := evpn4ovn.Spec.OvsdbService.Endpoints.Ports
	var svcPorts []corev1.ServicePort
	for i := range ports {
		svcPorts = append(svcPorts, corev1.ServicePort{
			Name:       ports[i].Name,
			Port:       ports[i].Port,
			TargetPort: intstr.FromInt32(ports[i].Port),
			Protocol:   corev1.ProtocolTCP,
			NodePort:   0,
		})
	}
	srv := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: evpn4ovn.Spec.OvsdbService.Name,
			//Namespace: evpn4ovn.ObjectMeta.Name,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			//ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			Ports: svcPorts,
		},
		Status: corev1.ServiceStatus{},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &srv, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create Service: %v", err)
	}
	return nil
}

func (r *Evpn4OvnReconciler) createOvsdbEndpoint(ctx context.Context, evpn4ovn *evpnapiv1alpha1.Evpn4Ovn) error {
	ports := evpn4ovn.Spec.OvsdbService.Endpoints.Ports
	var epPorts []corev1.EndpointPort
	for i := range ports {
		epPorts = append(epPorts, corev1.EndpointPort{
			Name: ports[i].Name,
			Port: ports[i].Port,
		})
	}
	ep := corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: evpn4ovn.Spec.RabbitmqService.Name,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: evpn4ovn.Spec.RabbitmqService.Endpoints.IP[0]},
				},
				Ports: epPorts,
			},
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &ep, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create Rabbitmq endpoint: %v", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Evpn4OvnReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&evpnapiv1alpha1.Evpn4Ovn{}).
		Complete(r)
}
