package htcjob

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path"

    htcv1alpha1 "gitlab.cern.ch/cms-cloud/htc-operator/pkg/apis/htc/v1alpha1"
    //corev1 "k8s.io/api/core/v1"
    //metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    //batchv1 "k8s.io/api/batch/v1"
)

func (r *ReconcileHTCJob) submitCondorJob(v *htcv1alpha1.HTCJob) ([]string, error) {
    tdName, err := ioutil.TempDir(os.TempDir(), "scratch-")
    queue_no := 1
    if v.Spec.Queue != 0 {
        queue_no = v.Spec.Queue
    }
    jobShellScript := "#!/bin/bash\n" +
        "singularity exec " +
        "       --bind /cvmfs " +
        "       --bind /afs/cern.ch " +
        "       --bind /eos " +
        fmt.Sprintf("docker://%s ", v.Spec.Script.Image) +
        "./script.sh\n" +
        "./sender $?\n" // send retcode too
    jobSubFile := "universe                = vanilla\n" +
        "executable              = job.sh\n" +
        "should_transfer_files   = Yes\n" +
        "when_to_transfer_output = ON_EXIT\n" +
        "output                  = out.$(ClusterId).$(ProcId)\n" +
        "error                   = err.$(ClusterId).$(ProcId)\n" +
        "log                     = log.$(ClusterId).$(ProcId)\n" +
        "transfer_input_files    = script.sh, /usr/local/bin/sender\n" +
        fmt.Sprintf("environment = \"JOB_NAME=%s TEMP_DIR=%s\"\n", v.Name, tdName) +
        fmt.Sprintf("queue %d\n", queue_no)
    // submit the job to HTC
    // write files
    if err != nil {
        fmt.Println("Failed tempdir creation")
        return nil, err
    }
    errmsg, err := sendJob(v.Spec.Script.Source, jobShellScript, jobSubFile, tdName)
    if err != nil {
        fmt.Println(errmsg)
        return nil, err
    }
    // record the submission in a database
    jobId, err := recordSubmission(v.Name, tdName)
    if err != nil {
        fmt.Print("Failed to record the fact of submission in the DB")
        return nil, err
    }
    return jobId, nil
}

func sendJob(script string, jobShellScript string, jobSubFile string,
    tempDirName string) (string, error) {
    err := ioutil.WriteFile(path.Join(tempDirName, "script.sh"),
        []byte(script), 0777)
    if err != nil {
        return "Failed writing script file", err
    }
    err = ioutil.WriteFile(path.Join(tempDirName, "job.sh"),
        []byte(jobShellScript), 0444)
    if err != nil {
        return "Failed writing job script file", err
    }
    err = ioutil.WriteFile(path.Join(tempDirName, "job.sub"),
        []byte(jobSubFile), 0444)
    if err != nil {
        return "Failed writing job submission file", err
    }
    // submit the job
    _, err = exec.Command("subCondor", tempDirName).CombinedOutput()
    if err != nil {
        return "Failed job submission", err
    }
    return "", nil
}
