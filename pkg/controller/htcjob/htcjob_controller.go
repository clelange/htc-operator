package htcjob

import (
    "context"
    "time"

    //"fmt"
    htcv1alpha1 "gitlab.cern.ch/cms-cloud/htc-operator/pkg/apis/htc/v1alpha1"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"

    //metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    //"k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller"

    //"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/manager"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    "sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_htcjob")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new HTCJob Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
    return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
    return &ReconcileHTCJob{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
    // Create a new controller
    c, err := controller.New("htcjob-controller",
        mgr, controller.Options{Reconciler: r})
    if err != nil {
        return err
    }

    // Watch for changes to primary resource HTCJob
    err = c.Watch(&source.Kind{Type: &htcv1alpha1.HTCJob{}},
        &handler.EnqueueRequestForObject{})
    if err != nil {
        return err
    }

    // TODO(user): Modify this to be the types you create that are owned by the primary resource
    // Watch for changes to secondary resource Pods and requeue the owner HTCJob
    err = c.Watch(&source.Kind{Type: &corev1.Pod{}},
        &handler.EnqueueRequestForOwner{
            IsController: true,
            OwnerType:    &htcv1alpha1.HTCJob{},
        })
    if err != nil {
        return err
    }

    return nil
}

// blank assignment to verify that ReconcileHTCJob implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileHTCJob{}

// ReconcileHTCJob reconciles a HTCJob object
type ReconcileHTCJob struct {
    // This client, initialized using mgr.Client() above, is a split client
    // that reads objects from the cache and writes to the apiserver
    client client.Client
    scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a HTCJob object and makes changes based on the state read
// and what is in the HTCJob.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHTCJob) Reconcile(request reconcile.Request) (
    reconcile.Result, error) {
    reqLogger := log.WithValues("Request.Namespace", request.Namespace,
        "Request.Name", request.Name)
    reqLogger.Info("Reconciling HTCJob")

    instance := &htcv1alpha1.HTCJob{}

    err := r.client.Get(context.TODO(), request.NamespacedName, instance)
    if err != nil {
        if errors.IsNotFound(err) {
            return reconcile.Result{}, nil
        }
        return reconcile.Result{}, err
    }
    // Check if the Job already exists
    if len(instance.Status.JobId) == 0 {
        // send the job and add an entry in the db
        // (after adding to active so many jobs dont get rescheduled)
        jobId, err := r.submitCondorJob(instance)
        if err != nil {
            reqLogger.Error(err, "Failed to send a job to HTCondor")
            return reconcile.Result{}, err
        }
        // record the jobId in Status
        instance.Status.JobId = jobId
        instance.Status.Active = len(jobId)
        //instance.Status.JobId = make([]string, len(jobId))
        //copy(instance.Status.JobId, jobId)
        err = r.client.Status().Update(context.TODO(), instance)
        if err != nil {
            reqLogger.Error(err, "Failed to update HTCJob status (Active)")
            return reconcile.Result{}, err
        }
        // Requeue to wait for the job to complete
        return reconcile.Result{RequeueAfter: time.Second * 10}, nil
    } else {
        // a job is active => check if it's marked as running in the database
        var everyJobStatus []int
        for _, currentJobId := range instance.Status.JobId {
            jobStatus, err := getJobStatus(instance.Name, currentJobId)
            if err != nil {
                reqLogger.Error(err, "Failed to get the status of an htcjob")
                return reconcile.Result{}, err
            }
            everyJobStatus = append(everyJobStatus, jobStatus)
        }
        // calculate number of active, succeeded
        instance.Status.Active = 0
        instance.Status.Succeeded = 0
        instance.Status.Failed = 0
        for _, s := range everyJobStatus {
            switch s {
            case 1:
                instance.Status.Active += 1
            case 4:
                instance.Status.Succeeded += 1
            case 7:
                instance.Status.Failed += 1
            }
        }
        err = r.client.Status().Update(context.TODO(), instance)
        if err != nil {
            reqLogger.Error(err, "Failed to update HTCJob status")
            return reconcile.Result{}, err
        }
        if instance.Status.Active == 0 {
            return reconcile.Result{}, nil
        }
        return reconcile.Result{RequeueAfter: time.Second * 10}, nil
    }

    return reconcile.Result{RequeueAfter: time.Second * 10}, nil
}
