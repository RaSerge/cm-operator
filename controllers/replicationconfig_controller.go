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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	configv1alpha1 "github.com/raserge/cm-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicationConfigReconciler reconciles a ReplicationConfig object
type ReplicationConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.scartel.dc,resources=replicationconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.scartel.dc,resources=replicationconfigs/status,verbs=get;update;patch

func (r *ReplicationConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("replicationconfig", req.NamespacedName)

	// Fetch the App instance.
	app := &configv1alpha1.ReplicationConfig{}
	err := r.Get(context.TODO(), req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new deployment.
	found := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: app.Spec.TargetName, Namespace: app.Spec.TargetNamespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define and create a new config.
			cm := r.configForApp(app)
			if err = r.Create(ctx, cm); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	// Ensure the data replicated
	keyApp := app.Namespace + app.Name

	for k, v := range app.Spec.Data {
		annotationKey := keyApp + k
		sha256b := sha256.Sum256([]byte(v))
		sha256s := hex.EncodeToString(sha256b[:])
		if _, ok := found.Annotations[annotationKey]; ok && v == sha256s {
			return ctrl.Result{Requeue: true}, nil
		} else {
			found.Annotations[annotationKey] = sha256s
			found.Data[annotationKey] = v
			err = r.Update(ctx, found)
			if err != nil {
				log.Error(err, "Failed to update Config", "Config.Namespace", found.Namespace, "Config.Name", found.Name)
				return ctrl.Result{}, err
			}
			// Spec updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}

	}

	// Update status.ReplicatedKeys if needed.
	keyNames := r.getAppDataKeys(app)
	if !reflect.DeepEqual(keyNames, app.Status.ReplicatedKeys) {
		app.Status.ReplicatedKeys = keyNames
		if err := r.Status().Update(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ReplicationConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.ReplicationConfig{}).
		Complete(r)
}

func (r *ReplicationConfigReconciler) getAppDataKeys(m *configv1alpha1.ReplicationConfig) []string {
	var keyNames []string
	for key := range m.Spec.Data {
		keyNames = append(keyNames, key)
	}
	return keyNames
}

func (r *ReplicationConfigReconciler) AnnotationsForConfig(m *configv1alpha1.ReplicationConfig) map[string]string {
	keyApp := m.Namespace + m.Name
	var keyValues map[string]string
	for k, v := range m.Spec.Data {
		annotationKey := keyApp + k
		keyValues[annotationKey] = v
	}
	return keyValues
}

// deploymentForApp returns a app Config object.
func (r *ReplicationConfigReconciler) configForApp(m *configv1alpha1.ReplicationConfig) *corev1.ConfigMap {

	as := r.AnnotationsForConfig(m)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Spec.TargetName,
			Namespace:   m.Spec.TargetNamespace,
			Annotations: as,
		},
		Data: m.Spec.Data,
	}

	// Set App instance as the owner and controller.
	// NOTE: calling SetControllerReference, and setting owner references in
	// general, is important as it allows deleted objects to be garbage collected.
	controllerutil.SetControllerReference(m, cm, r.Scheme)
	return cm
}
