#!/bin/bash

echo "apiVersion: v1
kind: Secret
metadata:
  name: kinit-secret
type: Opaque
data:
  username: `echo $OS_USERNAME|base64`
  password: `echo $OS_PASSWORD|base64`"| kubectl create -f -

kubectl create -f deploy/ceph.yaml
