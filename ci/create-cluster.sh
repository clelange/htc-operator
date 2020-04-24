#!/bin/bash
# create tha cluster
openstack coe cluster create ${CLUSTER_NAME} \
  --keypair lxplus \
  --cluster-template kubernetes-1.15.3-2 \
  --node-count 1 \
  --labels influx_grafana_dashboard_enabled=true \
  --labels cephfs_csi_enabled=true \
  --labels kube_csi_version=cern-csi-1.0-1 \
  --labels cloud_provider_tag=v1.15.0 \
  --labels container_infra_prefix=gitlab-registry.cern.ch/cloud/atomic-system-containers/ \
  --labels manila_enabled=true \
  --labels cgroup_driver=cgroupfs \
  --labels autoscaler_tag=v1.15.2 \
  --labels kube_csi_enabled=true \
  --labels flannel_backend=vxlan \
  --labels cvmfs_csi_version=v1.0.0 \
  --labels admission_control_list=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,Priority \
  --labels ingress_controller=traefik \
  --labels manila_version=v0.3.0 \
  --labels cvmfs_csi_enabled=true \
  --labels heat_container_agent_tag=stein-dev-2 \
  --labels kube_tag=v1.15.3 \
  --labels cephfs_csi_version=cern-csi-1.0-1

sleep 10

get_status () {
    openstack coe cluster show $CLUSTER_NAME -c status \
        | grep 'status ' \
        | awk '{print $4}'
}

STATUS=`get_status`
# wait till complete
while [ "$STATUS" != "CREATE_COMPLETE" ] && [ "$STATUS" != "CREATE_FAILED" ]
do
    STATUS=`get_status`
done
# get the cluster config (in order to access Kubernetes cluster)
openstack coe cluster show $CLUSTER_NAME
openstack coe cluster config $CLUSTER_NAME
export KUBECONFIG=config
# DNS configuration
kubectl label node $A_NODE role=ingress
export A_NODE=`kubectl get no|grep 'node-0'|awk '{print $1}'`
openstack server set --property landb-alias=$CLUSTER_NAME--load-1- $A_NODE
sleep 1000
