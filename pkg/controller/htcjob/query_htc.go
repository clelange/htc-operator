package htcjob

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type htcStatus struct {
	ClusterID            int    `json:"ClusterId"`
	ProcID               int    `json:"ProcId"`
	JobStatus            int    `json:"JobStatus"`
	ExitCode             int    `json:"ExitCode"`
	EnteredCurrentStatus int    `json:"EnteredCurrentStatus"`
	RemoveReason         string `json:"RemoveReason"`
}

func queryStatus(clusterID string) error {
	cmd := exec.Command("python", "query_htc.py", "--status", clusterID)
	log.Info("Querying job status using python API")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("StdoutPipe error: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("StdoutPipe error: %v", err)
	}
	var htcStatusList []htcStatus
	if err := json.NewDecoder(stdout).Decode(&htcStatusList); err != nil {
		return fmt.Errorf("Could not decode stdout: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("Wait error: %v", err)
	}
	fmt.Print(htcStatusList)
	return nil
}
