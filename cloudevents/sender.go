package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {

	ctx := cloudevents.ContextWithTarget(context.Background(),
		"http://cms-batch.cern.ch/cloudevents")
	//        "http://cms-batch.cern.ch")

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(),
		cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	// get job name
	jobName := os.Getenv("JOB_NAME")
	jobID := getJobID()

	e := cloudevents.NewEvent()
	e.SetType("htcjob.cloudevent")
	e.SetSource("cern.ch")
	_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"name":    jobName,
		"jobid":   jobID,
		"retcode": os.Args[1],
	})

	err = c.Send(ctx, e)
	if err != nil {
		log.Printf("failed to send: %v", err)
	}
}

func getJobID() string {
	var clusterID, procID string

	buf, err := os.Open(".job.ad")
	if err != nil {
		fmt.Println("File reading error")
		return "ERROR"
	}
	defer buf.Close()

	snl := bufio.NewScanner(buf)
	// The file .job.ad contains lines describing the jobID, e.g.:
	// ClusterId = 3974861
	// ProcId = 0
	reCluster := regexp.MustCompile(`^ClusterId = (.*)$`)
	reProc := regexp.MustCompile(`^ProcId = (.*)$`)
	for snl.Scan() {
		currText := snl.Text()
		if reCluster.MatchString(currText) {
			clusterID = reCluster.ReplaceAllString(currText, `$1`)
		}
		if reProc.MatchString(currText) {
			procID = reProc.ReplaceAllString(currText, `$1`)
		}
	}
	err = snl.Err()
	if err != nil {
		fmt.Println("File reading error")
		return "ERROR"
	}
	return fmt.Sprintf("%s.%s", clusterID, procID)
}
