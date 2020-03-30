package htcjob

import (
    "os"
    "os/exec"
    "path"
    "io/ioutil"
    "fmt"

    htcv1alpha1 "htc-operator/pkg/apis/htc/v1alpha1"
    //corev1 "k8s.io/api/core/v1"
    //metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    //batchv1 "k8s.io/api/batch/v1"
)

func (r *ReconcileHTCJob) submitCondorJob(v *htcv1alpha1.HTCJob) (string, error) {
    tdName, err := ioutil.TempDir(os.TempDir(), "scratch-")
    jobShellScript := "#!/bin/bash\n" +
        "ls -la\n" +
        "singularity exec --contain --ipc --pid " +
        "       --home $PWD:/srv " +
        "       --bind /cvmfs " +
        fmt.Sprintf("docker://%s ", v.Spec.Script.Image) +
        fmt.Sprintf("./script_%s.sh\n", v.Name) +
        "./sender\n"
    jobSubFile := "universe                = vanilla\n" +
        "executable              = job.sh\n" +
        "should_transfer_files   = Yes\n" +
        "when_to_transfer_output = ON_EXIT\n" +
        "output                  = out.$(ClusterId).$(ProcId)\n" +
        "error                   = err.$(ClusterId).$(ProcId)\n" +
        "log                     = log.$(ClusterId).$(ProcId)\n" +
        "transfer_input_files    = script.sh, sender\n" +
        fmt.Sprintf("environment = \"JOB_NAME=%s TEMP_DIR=%s\"\n", v.Name, tdName) +
        "queue 1"
    // submit the job to HTC
    // write files
    if err != nil {
        fmt.Println("Failed tempdir creation")
        return "", err
    }
    err = sendJob(v.Spec.Script.Source, jobShellScript, jobSubFile, tdName)
    if err != nil {
        fmt.Println("Failed job submission (function sendJob)")
        return "", err
    }
    // record the submission in a database
    jobId, err := recordSubmission(v.Name, tdName)
    if err != nil {
        fmt.Print("Failed to record the fact of submission in the DB")
        return "", err
    }
    return jobId, nil
}

func sendJob(script string, jobShellScript string, jobSubFile string,
    tempDirName string) error {
    err := ioutil.WriteFile(path.Join(tempDirName, "script.sh"),
        []byte(script), 0777)
    if err != nil {
        fmt.Println("Failed writing script file")
        return err
    }
    err = ioutil.WriteFile(path.Join(tempDirName, "job.sh"),
        []byte(jobShellScript), 0444)
    if err != nil {
        fmt.Println("Failed writing job script file")
        return err
    }
    err = ioutil.WriteFile(path.Join(tempDirName, "job.sub"),
        []byte(jobSubFile), 0444)
    if err != nil {
        fmt.Println("Failed writing job submission file")
        return err
    }
    // submit the job
    out, err := exec.Command("subCondor", tempDirName).CombinedOutput()
    if err != nil {
        fmt.Println("Failed job submission")
        return err
    }
    fmt.Printf("%s\n", out)
    return nil
}
