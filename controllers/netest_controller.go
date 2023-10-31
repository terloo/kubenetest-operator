/*
Copyright 2023.
*/

package controllers

import (
	"context"
	"net/netip"
	"time"

	netestv1alpha1 "github.com/terloo/kubenetest-operator/api/v1alpha1"
	"github.com/terloo/kubenetest-operator/controllers/utils"
	"github.com/terloo/kubenetest-operator/pkg/meta"
	"github.com/terloo/kubenetest-operator/pkg/worker"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var NetestNamespace string = "kubenetest"

// NetestReconciler reconciles a Netest object
type NetestReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Workers map[string]*worker.NetestWorker
}

//+kubebuilder:rbac:groups=netest.terloo.github.com,resources=netests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=netest.terloo.github.com,resources=netests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netest.terloo.github.com,resources=netests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Netest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *NetestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	netest := &netestv1alpha1.Netest{}
	err := r.Client.Get(ctx, req.NamespacedName, netest)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.InfoS("crd has deleted", "name", req.Name)
			r.Workers[req.Name].Close(req.Name)
			return ctrl.Result{}, nil
		}
		klog.Error(err, err.Error())
		return ctrl.Result{}, err
	}

	// create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: NetestNamespace},
	}
	err = r.Client.Get(ctx, client.ObjectKey{Name: NetestNamespace}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Client.Create(ctx, ns)
		}
		klog.Error(err, err.Error())
		return ctrl.Result{}, err
	}

	// creat daemonSet agent
	ds := &appsv1.DaemonSet{}
	ds, err = utils.RenderDaemonSet(*netest, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	controllerutil.SetOwnerReference(netest, ds, r.Scheme)

	currentDS := &appsv1.DaemonSet{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: NetestNamespace, Name: req.Name}, currentDS)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(ctx, ds)
		}
		if err != nil {
			klog.Error(err, err.Error())
			return ctrl.Result{}, err
		}
	}
	ds = currentDS

	// wait for ds ready
	if ds.Status.DesiredNumberScheduled == 0 || ds.Status.NumberReady != ds.Status.DesiredNumberScheduled {
		klog.Infof("ds %s is not ready...", ds.Name)
		return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil
	}

	// obtain pod of ds
	podList := &corev1.PodList{}
	err = r.Client.List(ctx, podList, client.InNamespace(ds.Namespace), client.MatchingLabels(map[string]string{
		"name": req.Name,
	}))
	if err != nil {
		return ctrl.Result{}, err
	}

	ipSlice := make([]*netip.Addr, len(podList.Items))
	for i, pod := range podList.Items {
		ip, err := netip.ParseAddr(pod.Status.PodIP)
		if err != nil {
			klog.Error(err, err.Error())
		}
		ipSlice[i] = &ip
	}

	w, ok := r.Workers[req.Name]
	if !ok {
		w = worker.NewNetestWorkers()
		r.Workers[req.Name] = w
	}

	for _, ip := range ipSlice {
		for _, targetIP := range ipSlice {
			if ip == targetIP {
				continue
			}

			work := &meta.NetestWork{
				Type:  meta.Ping,
				Value: targetIP.String(),
			}
			r.Workers[req.Name].Work(ip, work)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Workers == nil {
		r.Workers = make(map[string]*worker.NetestWorker)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&netestv1alpha1.Netest{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}
