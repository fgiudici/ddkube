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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/fgiudici/ddflare"
	ddkubev1beta1 "github.com/fgiudici/ddkube/api/v1beta1"
)

const (
	Cloudflare = "Cloudflare"
	Dyn        = "Dyn"
	NoIP       = "NoIP"
	DDNS       = "DDNS"
)

// HostnameReconciler reconciles a Hostname object
type HostnameReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ddkube.foggy.day,resources=hostnames,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ddkube.foggy.day,resources=hostnames/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ddkube.foggy.day,resources=hostnames/finalizers,verbs=update
//
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Hostname object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *HostnameReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	hostname := &ddkubev1beta1.Hostname{}
	if err := r.Get(ctx, req.NamespacedName, hostname); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	patch := client.MergeFrom(hostname.DeepCopy())

	// Get DDNSService AuthSecretRef
	authSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      hostname.Spec.DDNSService.AuthSecretRef.Name,
		Namespace: hostname.Spec.DDNSService.AuthSecretRef.Namespace,
	}, authSecret); err != nil {
		log.Error(err, "cannot retrieve DDNS service auth secret",
			"secretName", hostname.Spec.DDNSService.AuthSecretRef.Name,
			"secretNamespace", hostname.Spec.DDNSService.AuthSecretRef.Namespace)
		return ctrl.Result{}, err
	}
	authToken := ""
	if token, ok := authSecret.Data["authToken"]; !ok {
		err := fmt.Errorf("'authToken' not found")
		log.Error(err, "cannot retrieve DDNS service authentication token",
			"secretName", hostname.Spec.DDNSService.AuthSecretRef.Name,
			"secretNamespace", hostname.Spec.DDNSService.AuthSecretRef.Namespace)
		return ctrl.Result{}, err
	} else {
		authToken = string(token)
	}

	address := hostname.Spec.Address
	if address == "" {
		if ip, err := ddflare.GetPublicIP(); err != nil {
			return ctrl.Result{}, err
		} else {
			address = ip
		}
	}
	lastUpdate := hostname.Status.LastUpdate
	lastUpdate.ScheduledAt.Time = time.Now()
	lastUpdate.Address = address
	lastUpdate.Hostname = hostname.Spec.Hostname
	if err := ddflareFQDNUpdate(
		hostname.Spec.DDNSService.Endpoint,
		authToken,
		hostname.Spec.Hostname,
		address); err != nil {
		lastUpdate.Failed = true
	}

	if err := r.Status().Patch(ctx, hostname, patch); err != nil {
		return ctrl.Result{}, err
	}

	if *hostname.Spec.CheckIntervalMinutes == 0 {
		return ctrl.Result{Requeue: false}, nil
	}

	nextUpdate := time.Duration(*hostname.Spec.CheckIntervalMinutes) * time.Minute

	return ctrl.Result{RequeueAfter: nextUpdate}, nil
}

func ddflareFQDNUpdate(ep, authToken, fqdn, ip string) error {
	var dm *ddflare.DNSManager
	var err error

	switch ep {
	case Cloudflare:
		dm, err = ddflare.NewDNSManager(ddflare.Cloudflare)
	case Dyn:
		dm, err = ddflare.NewDNSManager(ddflare.Dyn)
	case NoIP:
		dm, err = ddflare.NewDNSManager(ddflare.NoIP)
	case DDNS:
		dm, err = ddflare.NewDNSManager(ddflare.DDNS)
	default:
		dm, err = ddflare.NewDNSManager(ddflare.Dyn)
		dm.SetApiEndpoint(ep)
	}
	if err != nil {
		return err
	}

	if err := dm.Init(authToken); err != nil {
		return err
	}

	isUpToDate := false
	if isUpToDate, err = dm.IsFQDNUpToDate(fqdn, ip); err != nil {
		return err
	}
	if isUpToDate {
		return nil
	}

	if err = dm.Update(fqdn, ip); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HostnameReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ddkubev1beta1.Hostname{}).
		Complete(r)
}
