package htcjob

import (
    //"context"
    //"time"
    "fmt"
    //"strings"

    htcv1alpha1 "htc-operator/pkg/apis/htc/v1alpha1"

    //appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    batchv1 "k8s.io/api/batch/v1"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    //"k8s.io/apimachinery/pkg/types"
    //"k8s.io/apimachinery/pkg/util/intstr"
    //"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    //"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileHTCJob) condorSubmitJob(v *htcv1alpha1.HTCJob, job_name string) *batchv1.Job {
    job_sub := "dW5pdmVyc2UgPSB2YW5pbGxhCmV4ZW" +
        "N1dGFibGUgICAgICAgICAgICAgID0gam9iLn" +
        "NoCm91dHB1dCAgICAgICAgICAgICAgICAgID" +
        "0gb3V0LnR4dApxdWV1ZSAxCg=="
    job_sh := "IyEvYmluL2Jhc2gKZWNobyBIRUxMT1dPUkxECg=="
    //fmt.Println(fmt.Sprintf("  echo '%s' |base64 --decode > job.sub &&", job_sub))
    //fmt.Println(fmt.Sprintf("  echo '%s' |base64 --decode > job.sh &&", job_sh))
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name: v.Name + "-job",
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
                            fmt.Sprintf("  echo '%s' |base64 --decode > job.sub && ", job_sub) +
                            fmt.Sprintf("  echo '%s' |base64 --decode > job.sh && ", job_sh) +
                            "  condor_submit -spool -verbose job.sub > condor.out && " +
                            "  cat condor.out && " +
                            "  s3cmd -c /mnt/s3cfg/..data/.s3cfg put condor.out" +
                            fmt.Sprintf("    s3://TADO_BUCKET/run_%s_condor.out'", job_name)},
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

func updateS3Objects(bucket string, job_name string) bool {
    accessKey := "PJE22MGQ8QO45CKI496K"
    secretKey := "8BnRJcxuOain8fLBxT44IfFMNgku7huHoo8mK8NP"
    s3Config := &aws.Config{
        Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
        Endpoint:         aws.String("s3.cern.ch"),
        Region:           aws.String("us-east-1"),
        DisableSSL:       aws.Bool(true),
        S3ForcePathStyle: aws.Bool(true),
    }
    svc := s3.New(session.New(s3Config))
    input := &s3.ListObjectsV2Input{
        Bucket:  aws.String(bucket),
        MaxKeys: aws.Int64(10000),
    }
    result, err := svc.ListObjectsV2(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            case s3.ErrCodeNoSuchBucket:
                fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return false
    }
    filenameInSlice := false
    for _, item := range result.Contents {
        if fmt.Sprintf("complete_%s_condor.out", job_name) == *(item.Key) {
            filenameInSlice = true
            input := &s3.DeleteObjectInput{
                Bucket: aws.String(bucket),
                Key:    aws.String(*(item.Key)),
            }
            _, err := svc.DeleteObject(input)
            if err != nil {
                if aerr, ok := err.(awserr.Error); ok {
                    switch aerr.Code() {
                    default:
                        fmt.Println(aerr.Error())
                    }
                } else {
                    fmt.Println(err.Error())
                }
                return false
            }
        }
    }
    return filenameInSlice
}
