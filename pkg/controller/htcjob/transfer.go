package htcjob

import (
	"fmt"
	"os/exec"
)

func (r *ReconcileHTCJob) transferCondorJob(jobID string) error {

	for tries := 1; true; tries++ {
		out, err := exec.Command("transferCondor", jobID).CombinedOutput()
		if err != nil && tries >= 2 {
			return fmt.Errorf("Transferring output for %v failed: %v", jobID, err)
		}
		if err == nil {
			fmt.Printf("Transferred output for %v: %v", jobID, string(out))
			break
		}
		fmt.Errorf("ERROR in transferCondor, trying again: %v", err)
	}

	return nil
}
