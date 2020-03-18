# HTCondor controller with operator-sdk

## Instructions

To create a CronJob object that will keep on querying HTC for
running jobs (if there are any files matching `s3://TADO_BUCKET/run_*`):

```
kubectl create -f watcher/config.yaml 
```

To create the controller, CRD and a sample CR:

```
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