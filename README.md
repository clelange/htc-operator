# HTCondor controller with operator-sdk

## Instructions

First of all, a couple of secrets are needed:
- a `gitlab-registry` secret with access to gitlab-registry.cern.ch
- a `kinit-secret` containing a username and password for the job submittion with `condorsubmit` container

To create a CronJob object that will keep on querying HTC for
running jobs (if there are any files matching `s3://TADO_BUCKET/run_*`):

```
kubectl create -f watcher/config.yaml 
```

To create the 
- controller
- CRD
- cloudevents watcher (service, deployment, ingress)
- a sample CR
:
```
# need to be created only once
kubectl create -f deploy/role_binding.yaml 
kubectl create -f deploy/service_account.yaml 
kubectl create -f deploy/role.yaml 
# script to be run after modifications to recreate resources
./recreate.sh
```

To check for running HTCJobs:

```
kubectl get htcjobs.htc.cern.ch
```

To get their status:

```
kubectl describe htcjobs.htc.cern.ch
```

Where under `Status`, number `Active` should should switch to `Succeeded` when the job completes in HTC.
