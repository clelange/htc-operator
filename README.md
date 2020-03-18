# HTCondor controller with operator-sdk

## Instructions

To create a CronJob object that will keep on querying HTC for
running jubs (if there are any files matching `s3://TADO_BUCKET/run_*`):

```
kubectl create -f watcher/config.yaml 
```

To create the controller, CRD and a sample CR:

```
./recreate
```

To check for running HTCJobs:

```
kubectl get htcjobs.htc.cern.ch
```

To ge the status:

```
kubectl describe htcjobs.htc.cern.ch
```

Where the number `Active` should should switch to `Succeeded` when the job completes in HTC.