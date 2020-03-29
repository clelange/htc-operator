# HTCondor controller with operator-sdk

## Instructions

### Secrets

First of all, a couple of secrets are needed:
- a `gitlab-registry` secret with access to gitlab-registry.cern.ch
- a `kinit-secret` containing a username and password for the job submittion with `condorsubmit` container

#### Gitlab registry secret

To create `gitlab-registry` secret to access the `condorsubmit` image, run:

```
kubectl create secret --namespace=default \
                      docker-registry gitlab-registry \
                      --docker-server=gitlab-registry.cern.ch \
                      --docker-username=<gitlab/cern username> \
                      --docker-password=<gitlab authentication token> \
                      --docker-email=<cern email> \
                      --output yaml --dry-run 
```

To be able to submit jobs from the container, it has to have access to a CERN user credentials: a username and a password.
To make `kinit- secret`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kinit-secret
type: Opaque
data:
  username: <username in base64>
  password: <password in base64>
```

### Database and storage

Job status is saved in a Postgresql database which uses a Ceph volume for storage.
To create the database deployment with its storage definitions, run:

```
kubectl create -f database/config.yaml 
```

To connect to the database, run:

```
psql -h cms-batch-test.cern.ch -U postgres -p 30303
```

The password is `pgpasswd`, the table - `htcjobs`.

### The Operator

The operator pod is based on a Docker image defined in `build/Dockerfile`,
which is generated from `build/Dockerfile.template`, so all modifications should be
done to the template.

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
kubectl get htcjobs
```

To get their status:

```
kubectl get htcjobs -o yaml
```

Where under `Status`, number `Active` should should switch to `Succeeded` when the job completes in HTC.
