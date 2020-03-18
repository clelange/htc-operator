#operator-sdk generate k8s
operator-sdk build xkxgygmoqkguuddnkz/htc-operator || exit
docker push xkxgygmoqkguuddnkz/htc-operator

kubectl create configmap s3cfg --from-file=$HOME/.s3cfg

kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/crds/htc.cern.ch_htcjobs_crd.yaml

kubectl create -f deploy/operator.yaml
kubectl create -f deploy/crds/htc.cern.ch_htcjobs_crd.yaml
kubectl create -f deploy/crds/htc.cern.ch_v1alpha1_htcjob_cr.yaml
