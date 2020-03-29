#operator-sdk generate k8s
# build the cloudevents receiver
sed "s/<YOUR_USER_NAME>/$USER/g" build/Dockerfile.template > build/Dockerfile
go build -o build/bin/receiver cloudevents/receiver.go
s3cmd put build/bin/receiver s3://TADO_BUCKET/receiver
# build the cloudevents sender
go build -o build/bin/sender cloudevents/sender.go
s3cmd put build/bin/sender s3://TADO_BUCKET/sender
operator-sdk build xkxgygmoqkguuddnkz/htc-operator || exit
docker push xkxgygmoqkguuddnkz/htc-operator

kubectl create configmap s3cfg --from-file=$HOME/.s3cfg

kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/crds/htc.cern.ch_htcjobs_crd.yaml

kubectl create -f deploy/operator.yaml
kubectl create -f deploy/crds/htc.cern.ch_htcjobs_crd.yaml
kubectl create -f deploy/crds/htc.cern.ch_v1alpha1_htcjob_cr.yaml
