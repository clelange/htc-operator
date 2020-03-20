package htcjob

import (
    //"context"
    //"time"
    "fmt"
    htcv1alpha1 "htc-operator/pkg/apis/htc/v1alpha1"

    //appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    batchv1 "k8s.io/api/batch/v1"
    //"k8s.io/apimachinery/pkg/types"
    //"k8s.io/apimachinery/pkg/util/intstr"
    //"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    //"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileHTCJob) transferCondorData(v *htcv1alpha1.HTCJob) *batchv1.Job {
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name: v.Name + "-transfer-job",
            Namespace: v.Namespace,
            Labels: map[string]string{
                "app": v.Name,
            },
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    ImagePullSecrets: []corev1.LocalObjectReference{{
                        Name: "gitlab-registry",
                    }},
                    Containers: []corev1.Container{{
                        Name: "condor-submission-example",
                        Image: "gitlab-registry.cern.ch/clange/condorsubmit",
                        Command: []string{"sh", "-c"},
                        Args: []string{
                            "export CONDOR_USER=`cat /mnt/kinit/username` && " +
                            "echo $CONDOR_USER && " +
                            "useradd -Ms /bin/bash $CONDOR_USER && " +
                            "mkdir -m 777 /scratch && " +
                            "mkdir -m 777 /home/tbareiki && " +
                            "pip install s3cmd && " +
                            "runuser -l $CONDOR_USER -c ' " +
                            "  cat /mnt/kinit/password| kinit && " +
                            "  cd /scratch && " +
                            "  export _condor_SCHEDD_HOST=bigbird10.cern.ch && " +
                            "  export _condor_CREDD_HOST=bigbird10.cern.ch && " +
                            fmt.Sprintf("  s3cmd -c /mnt/s3cfg/..data/.s3cfg get s3://TADO_BUCKET/complete_%s_condor.out && ", v.Name) +
                            fmt.Sprintf("  for i in `cat complete_%s_condor.out`; do condor_transfer_data $i; done && ", v.Name) +
                            fmt.Sprintf("  s3cmd -c /mnt/s3cfg/..data/.s3cfg rm s3://TADO_BUCKET/complete_%s_condor.out && ", v.Name) +
                            fmt.Sprintf("  s3cmd -c /mnt/s3cfg/..data/.s3cfg rm s3://TADO_BUCKET/script_%s.sh && ", v.Name) +
                            fmt.Sprintf("  rm complete_%s_condor.out && ", v.Name) +
                            fmt.Sprintf("  tar cvzf results_%s.tar.gz ./* && ", v.Name) +
                            fmt.Sprintf("  s3cmd -c /mnt/s3cfg/..data/.s3cfg put results_%s.tar.gz s3://TADO_BUCKET/ '", v.Name)},
                        VolumeMounts: []corev1.VolumeMount{
                            {
                                Name: "kinit-secret-vol",
                                MountPath: "/mnt/kinit",
                            },
                            {
                                Name: "s3cfg-vol",
                                MountPath: "/mnt/s3cfg",
                            },
                        },
                    }},
                    Volumes: []corev1.Volume{
                        {
                            Name: "kinit-secret-vol",
                            VolumeSource: corev1.VolumeSource{
                                Secret: &corev1.SecretVolumeSource{
                                    SecretName: "kinit-secret",
                                },
                            },
                        },
                        {
                            Name: "s3cfg-vol",
                            VolumeSource: corev1.VolumeSource{
                                ConfigMap: &corev1.ConfigMapVolumeSource{
                                    LocalObjectReference: corev1.LocalObjectReference{
                                        Name: "s3cfg",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    return job
}
