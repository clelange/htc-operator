#!/bin/bash
set -x
echo "apiVersion: v1
kind: Secret
metadata:
  name: kinit-secret
type: Opaque
data:
  username: `echo $OS_USERNAME|base64`
  password: `echo $OS_PASSWORD|base64`"| kubectl create -f -

kubectl create -f deploy/ceph.yaml

kubectl create secret \
  docker-registry gitlab-registry \
  --docker-server=gitlab-registry.cern.ch \
  --docker-username=$COE_USER \
  --docker-password=$REGISTRY_TOK \
  --docker-email=$GITLAB_USER_EMAIL \
  --output yaml --dry-run| kubectl create -f -
