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
	tdName, err := ioutil.TempDir("/data/tmp_jobs", "scratch-")
	// create the tempdir
	err = os.MkdirAll(tdName, 0777)
	queueNo := 1
	if v.Spec.Queue != 0 {
		queueNo = v.Spec.Queue
	}
	jobShellScript := "#!/bin/bash\n" +
		"singularity exec " +
		"       --bind /cvmfs " +
		"       --bind /afs/cern.ch " +
		"       --bind /eos " +
		fmt.Sprintf(" docker://%s ", v.Spec.Script.Image) +
		fmt.Sprintf(" %s ./script.sh $1\n", v.Spec.Script.Command) +
		"./sender $?\n" // send retcode too
	jobSubFile := "universe                = vanilla\n" +
		"executable              = job.sh\n" +
		"+MaxRuntime = 1200\n" +
		"+AccountingGroup = \"group_u_CMST3.all\"\n" +
		"arguments               = $(ProcId)\n" +
		"should_transfer_files   = Yes\n" +
		"when_to_transfer_output = ON_EXIT\n" +
		"output                  = out.$(ClusterId).$(ProcId)\n" +
		"error                   = err.$(ClusterId).$(ProcId)\n" +
		"log                     = log.$(ClusterId).$(ProcId)\n" +
		"transfer_input_files    = script.sh, /usr/local/bin/sender\n" +
		fmt.Sprintf("environment = \"JOB_NAME=%s TEMP_DIR=%s\"\n", v.Name, tdName) +
		fmt.Sprintf("\n%s\n", v.Spec.HTCopts) +
		fmt.Sprintf("queue %d\n", queueNo)
	// submit the job to HTC
	// write files
	errmsg, err := sendJob(v.Spec.Script.Source, jobShellScript, jobSubFile, tdName)
	if err != nil {
		fmt.Println(errmsg)
		return nil, err
	}
	// record the submission in a database
	jobID, err := recordSubmission(v.Name, tdName, v.Status.UniqId)
	if err != nil {
		fmt.Print("Failed to record the fact of submission in the DB")
		return nil, err
	}
	return jobID, nil
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
	// try twice
	for tries := 1; true; tries++ {
		out, err := exec.Command("subCondor", tempDirName).CombinedOutput()
		if err != nil && tries >= 2 {
			return "Failed job submission", err
		}
		if err == nil {
			fmt.Println(string(out))
			break
		}
		fmt.Println("ERROR in send, subCondor")
		fmt.Println(err)
	}
	return "", nil
}
