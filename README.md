# HTCondor Job operator

## Introduction

This is a Kubernetes operator that allows the submission of HTCondor jobs.
When the operator is installed in the cluster,
it creates a Pod which periodically checks the status of Kubernetes HTCJob resources
and updates their status based on the HTCondor jobs associated with them.

## Setup

The operator uses several external items that need to be created in order for it to work.

To use HTCondor, CERN user credentials, a username and a password, are needed.
They are extracted from a secret named `kinit-secret` which looks like the following:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kinit-secret
type: Opaque
data:
  username: <username in BASE64>
  password: <password in BASE64>
```

An `sqlite3` database is used to record the statuses of HTCJobs. It is stored in a `cephfs` volume, so to access it, 
StorageClass `csi-cephfs-cms` definition is needed, which looks something like:

```
apiVersion: storage.k8s.io/v1beta1
kind: StorageClass
metadata:
  name: csi-cephfs-cms
provisioner: manila-provisioner
parameters:
  type: "Geneva CephFS Testing"
  zones: nova
  osSecretName: os-trustee
  osSecretNamespace: kube-system
  protocol: CEPHFS
  backend: csi-cephfs
  csi-driver: cephfs.csi.ceph.com
  # the share should be of size > 1 
  # id for some_share from `manila list`
  osShareID: <ID>
  # id from `manila access-list some_share`
  osShareAccessID: <ID>
```


Only the two marked fields need to be changed. Access to these Ids should be limited,
as they allow the modification of the database that is stored in the `cephfs` volume.
More instructions on `cephfs` and `manila` can be found [here](https://clouddocs.web.cern.ch/containers/tutorials/cephfs.html)(Kubernetes)
and [here](https://clouddocs.web.cern.ch/file_shares/quickstart.html)(manila).

__In addition, the volume is assumed to have a database stored in a file in `/sqlite/htcjobs.db`__
__The folder in which the database file is stores must be writeable, not only the database file__
TODO: automate the creation of the database file if it doesn't exist.

The database holds a table named `htcjobs` created with:

```
# run in shell to create the database
sqlite3 htcjob.db
# run in sqlite to create the table
create table htcjobs(
    htcjobName varchar,
    jobId char(10),
    status integer,
    tempDir varchar
);
```

## Building the operator

The operator pod uses a Docker image based on `build/Dockerfile`.
The user inside the container must be the same as in the `kinit-secret`, 
so the following line has to be updated:

```
ARG USER_NAME=<your_username>
```

Then two separate executables have to be compiled into a directory
that will be moved to the container:

```
go build -o build/bin/receiver cloudevents/receiver.go
go build -o build/bin/sender cloudevents/sender.go
```

To compile the rest of the code, build the container image and
push it to Docker Hub, run

```
operator-sdk build xkxgygmoqkguuddnkz/htc-operator
docker push xkxgygmoqkguuddnkz/htc-operator
```
The `xkxgygmoqkguuddnkz` repository is chosen at 'random' and should be changed.

These steps should be repeated after changes to code are made.

## Operator deployment

To create the HTCJob CRD and operator, run:

```
kubectl create -f deploy/operator.yaml
kubectl create -f deploy/crds/htc.cern.ch_htcjobs_crd.yaml
```

## Example

A simple example can be found in `examples/retcode/zero.yaml`:

```yaml
apiVersion: htc.cern.ch/v1alpha1
kind: HTCJob
metadata:
  name: zero
spec:
  name: zero
  script:
    image: centos
    command: bash
    source: |
      echo "A"
```

When this HTCJob is created in the cluster, its current state
can be obtained with

```
kubectl get htcjob zero -o yaml
```

When the job is sent to HTCondor, the status of the HTCJob is updated and
the description shows and additional `status` field

```yaml
status:
  active: 1
  failed: 0
  jobid:
  - "884426.0"
  succeeded: 0
```

which specifies the job status and Id in HTCondor.

## HTCJob resource fields

The operator submits an HTCondor job, which executes `singularity` with arguments
taken from the HTCJob spec:
```
.spec.script.image: container image to use
.spec.script.command: command to be run in the container
.spec.script.source: contents of a script file that is given as an argument to `.spec.script.command'
.spec.script.queue: optional, specify the number of jobs to be sent. The job number in the sequence is an argument to the script
```
After the job is submitted, its `.status` field gets populated:
```
.status.active: number of  currently running HTCondor jobs
.status.failed: number of HTCondor jobs in which `singularity` exited with a return code != 0
.status.succeeded: number of HTCondor jobs in which `singularity` exited with a return code == 0
.status.jobid: array of Ids of the jobs submitted to HTCondor
```
## Logic and components of the operator

The operator is composed of three parts:

- operator application
- sqlite database
- cloudevents receiver and listener

## What happens with each HTCJob step-by-step

### 1. Job submission

The htc-operator pod runs a listener application `build/_output/bin/htc-operator`.
When an HTCJob is created in the cluster, the operator application reads its `.spec` and
submits an HTCondor job based on it.

### 2. Recording the status of the job

After the job is submitted, its status is recorded in two places:
- `.status` field of the HTCJob resource in the Kubernetes cluster
- sqlite database

The following values are written to the table `htcjobs`.

- htcjobName:name of the HTCJob resource
- jobId: job Id (ClusterId.ProcId)
- status: job status, initial value of which is set to '1', but later is changed to '4' or '7'
- tempDir: path from which the job was sent (note on this at 'Retrieving the logs from HTCondor')

### 3. Listening

After the submission of the job is recorded, the operator application runs a loop
(the Reconcile loop) that checks whether all jobs submitted have 'succeeded'.
The database is queried every 10s whether all the jobs from
`.status.jobid` have the status value equal to '4' in the table `htcjobs`.

### 4. Running the HTCondor job

The HTCondor job runs the user-defined script in a `singularity` container.
Then, when the `singularity` process exits, its return code is recorded.
The return code and HTCJob name are given as arguments to the `sender` application,
which is run from within the HTCondor job.

### 5. Updating the job status

Alongside the `htc-operator` application, in the operator pod runs a `receiver` application.
When the HTCondor job finishes, the 'receiver' gets an HTTP request from the `sender`, which includes
the HTCJob name, jobid and return code, which are used to update the HTCondor job status in the database.
If the return code is equal to zero, status value in `htcjobs` table is set to `4` for that job, `7` otherwise.
The `htc-operator` application calculates the number of completed jobs during every query. If the status is equal to `4`,
it contributes to the count of `.status.succeeded`, if it is equal to `7` - to `.status.failed`,
and if it is hasn't changed and is still equal to `1` - to `.status.active`.

### 6. Finishing the Reconcile loop

When the number of `.status.active` becomes equal to zero, that is, all HTCJobs have completed,
the Reconcile loop exits and the state of the HTCJob resource is no longer modified.

# Retrieving logs from HTCondor

The jobs are submitted with `condor_submit -spool`, since the `-spool` option allows
manual retrieval of logs of the completed job. In order to get these logs, `condor_transfer_data`
can be used. However, the data can only be transferred to the same directory from which the job was submitted.
The directory name is saved in the database and also is outputted by `condor_transfer_data` in case of error.

# Resources used

Openshift tutorial for operator-sdk:
[here](https://docs.openshift.com/container-platform/4.2/operators/operator_sdk/osdk-getting-started.html)

General overview of Kubernetes operators and their frameworks:
[here](https://www.oreilly.com/library/view/kubernetes-operators/9781492048039/)

Metacontroller - a simpler operator framework that helped me get a better understanding
of Kubernetes controllers/operators: [here](https://metacontroller.app/)