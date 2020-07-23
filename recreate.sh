#!/bin/sh
# operator-sdk generate k8s
go build -o build/bin/receiver cloudevents/receiver.go
go build -o build/bin/sender cloudevents/sender.go
# operator-sdk build gitlab-registry.cern.ch/cms-cloud/htc-operator --image-build-args "--build-arg CI_PROJECT_NAMESPACE=cms-cloud --build-arg CI_PROJECT_NAME=htc-operator" || exit
CGO_ENABLED=1 GOOS=linux operator-sdk build gitlab-registry.cern.ch/cms-cloud/htc-operator|| exit
docker push gitlab-registry.cern.ch/cms-cloud/htc-operator || exit

kubectl create configmap s3cfg --from-file="${HOME}/.s3cfg"

kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/crds/htc.cern.ch_htcjobs_crd.yaml

kubectl create -f deploy/operator.yaml
kubectl create -f deploy/crds/htc.cern.ch_htcjobs_crd.yaml
kubectl create -f deploy/crds/htc.cern.ch_v1alpha1_htcjob_cr.yaml
