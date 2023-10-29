/*
Copyright 2023.
*/

package controllers

import (
	"context"
	"time"

	netestv1alpha1 "github.com/terloo/kubenetest-operator/api/v1alpha1"
	"github.com/terloo/kubenetest-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NetestReconciler reconciles a Netest object
type NetestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	clog := log.FromContext(ctx)

	errResult := ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}

	netest := &netestv1alpha1.Netest{}
	err := r.Client.Get(ctx, req.NamespacedName, netest)
	if err != nil {
		return errResult, client.IgnoreNotFound(err)
	}

	// create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "kubenetest"},
	}
	err = r.Client.Get(ctx, req.NamespacedName, ns)
	if errors.IsNotFound(err) {
		r.Client.Create(ctx, ns)
	} else if err != nil {
		clog.Error(err, err.Error())
	}

	// creat daemonSet agent
	ds := &appsv1.DaemonSet{}
	err = r.Client.Get(ctx, req.NamespacedName, ds)
	if errors.IsNotFound(err) {
		ds, err = utils.RenderDaemonSet(*netest, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.SetOwnerReference(netest, ds, r.Scheme)
		err = r.Client.Create(ctx, ds)
		if err != nil {
			clog.Error(err, err.Error())
			return ctrl.Result{}, nil
		}
	} else if err != nil {
		clog.Error(err, err.Error())
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netestv1alpha1.Netest{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}
