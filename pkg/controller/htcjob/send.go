package htcjob

import (
    //"context"
    //"time"
    "fmt"
    "strings"
    "encoding/base64"
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

func (r *ReconcileHTCJob) condorSubmitJob(v *htcv1alpha1.HTCJob) *batchv1.Job {
    job_sh := "#!/bin/bash\n" +
        "ls -la\n" +
        "singularity exec --contain --ipc --pid " +
        "       --home $PWD:/srv " +
        "       --bind /cvmfs " +
        fmt.Sprintf("docker://%s ", v.Spec.Container) +
        fmt.Sprintf("./script_%s.sh", v.Name)
    job_sub := "universe                = vanilla\n" +
        "executable              = job.sh\n" +
        "should_transfer_files   = IF_NEEDED\n" +
        "when_to_transfer_output = ON_EXIT\n" +
        "output                  = out.$(ClusterId).$(ProcId)\n" +
        "error                   = err.$(ClusterId).$(ProcId)\n" +
        "log                     = log.$(ClusterId).$(ProcId)\n" +
        fmt.Sprintf("transfer_input_files    = script_%s.sh\n", v.Name) +
        "queue 1"
    // put the script to s3
    accessKey := "PJE22MGQ8QO45CKI496K"
    secretKey := "8BnRJcxuOain8fLBxT44IfFMNgku7huHoo8mK8NP"
    s3Config := &aws.Config{
        Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
        Endpoint:         aws.String("s3.cern.ch"),
        Region:           aws.String("us-east-1"),
        DisableSSL:       aws.Bool(true),
        S3ForcePathStyle: aws.Bool(true),
    }
    bucket := "TADO_BUCKET"
    svc := s3.New(session.New(s3Config))
    input := &s3.PutObjectInput{
        Body:                 aws.ReadSeekCloser(strings.NewReader(v.Spec.Script)),
        Bucket:               aws.String(bucket),
        Key:                  aws.String(fmt.Sprintf("script_%s.sh", v.Name)),
    }
    result, err := svc.PutObject(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return nil
    }
    fmt.Println(result)
    // create the job
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
                            fmt.Sprintf("  echo '%s' |base64 --decode > job.sub && ", base64.StdEncoding.EncodeToString([]byte(job_sub))) +
                            fmt.Sprintf("  echo '%s' |base64 --decode > job.sh && ", base64.StdEncoding.EncodeToString([]byte(job_sh))) +
                            fmt.Sprintf("  s3cmd -c /mnt/s3cfg/..data/.s3cfg get s3://TADO_BUCKET/script_%s.sh && ", v.Name) +
                            fmt.Sprintf("  chmod 777 script_%s.sh && ", v.Name) +
                            "  condor_submit -spool -verbose job.sub > condor.out && " +
                            "  cat condor.out && " +
                            "  s3cmd -c /mnt/s3cfg/..data/.s3cfg put condor.out" +
                            fmt.Sprintf("    s3://TADO_BUCKET/run_%s_condor.out'", v.Name)},
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
