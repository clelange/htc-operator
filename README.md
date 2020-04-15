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

A special Docker image is used to submit jobs to HTCondor.
To create a secret `gitlab-registry`, which is  needed
to pull the `clange/condorsubmit` image from CERN's gitlab registry, run:

```
kubectl create secret --namespace=default \
    docker-registry gitlab-registry \
    --docker-server=gitlab-registry.cern.ch \
    --docker-username=<gitlab/cern username> \
    --docker-password=<gitlab authentication token> \
    --docker-email=<cern email> \
    --output yaml --dry-run
```
Taken from [here](https://blog.zedroot.org/2019/01/21/gitlab-ci-kubernetes-pull-a-private-image-from-a-k8s-pod/).


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

Only the two marked fields need to be changed.
Taken from [here](https://clouddocs.web.cern.ch/containers/tutorials/cephfs.html).

__In addition, the volume is assumed to have a database stored in a file in `/sqlite/htcjobs.db`__

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

When the job is sent to HTCondor, the description gains an additional `status` field

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

- .spec.script.image: container image to use
- .spec.script.command: command to be run in the container
- .spec.script.source: contents of a script file that is given as
an argument to `.spec.script.command'
- .spec.script.queue: optional, specify the number of jobs to be sent.
The job number in the sequence is an argument to the script

After the job is submitted, its `.status` field gets populated:

- .status.active: number of  currently running HTCondor jobs
- .status.failed: number of HTCondor jobs in which `singularity` exited with a return code != 0
- .status.succeeded: number of HTCondor jobs in which `singularity` exited with a return code == 0
- .status.jobid: array of Ids of the jobs submitted to HTCondor

## Logic and components of the operator

The operator is composed of three parts:

- operator application
- sqlite database
- cloudevents receiver and listener

### htc-operator

This is the program that runs in its own pod and updates the state of each HTCJob resource.
It submits a job to HTCondor based on the resource specification and records the HTC job Id
and status in the database. Then it waits for the status value to change in the
database to update the status of the resource in the Kubernetes cluster.

The main code for the operator is in `pkg/controller/htcjob/`.
The operator deployment specification, along with additional resources is defined in
`deploy/operator.yaml`.

## sqlite database

The `sqlite` database is composed of a single file that is held in a `ceph` volume.
To make a switch to `postgresql`, some minimal changes have to be made to
`pkg/controller/htcjob/db.go`, and an example deployment
for the database server is in `database/config.yaml`.

## cloudevents receiver and listener

Alongside the main htc-operator program in the operator pod runs a cloudevents
receiver. Also, each HTCondor job, after the singularity process exits,
runs a cloudevents sender executable that sends an HTTP request to the
receiver with the job Id and the return code. Then the receiver updates the job
status based on the job Id in the sqlite database.

Code for both executables is located in `cloudevents/sender.go` and `cloudevents/receiver.go`.

# Resources used

Openshift tutorial for operator-sdk:
[here](https://docs.openshift.com/container-platform/4.2/operators/operator_sdk/osdk-getting-started.html)

General overview of Kubernetes operators and their frameworks:
[here](https://www.oreilly.com/library/view/kubernetes-operators/9781492048039/)

Metacontroller - a simpler operator framework that helped me get a better understanding
of Kubernetes controllers/operators: [here](https://metacontroller.app/)
