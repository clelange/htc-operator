package main

import (
    "context"
    "log"
    "fmt"
    "os"
    "bufio"
    "regexp"
    cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {

    ctx := cloudevents.ContextWithTarget(context.Background(),
        "http://cms-batch-test.cern.ch")

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
    jobId := getJobId()

    e := cloudevents.NewEvent()
    e.SetType("htcjob.cloudevent")
    e.SetSource("cern.ch")
    _ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
        "name": jobName,
        "jobid": jobId,
        "retcode": os.Args[1],
    })

    err = c.Send(ctx, e)
    if err != nil {
        log.Printf("failed to send: %v", err)
    }
}

func getJobId() string {
    var clusterId, procId string

    buf, err := os.Open(".job.ad")
    if err != nil {
        fmt.Println("File reading error")
        return "ERROR"
    }
    defer buf.Close()

    snl := bufio.NewScanner(buf)
    reCluster := regexp.MustCompile(`^ClusterId = (.*)$`)
    reProc := regexp.MustCompile(`^ProcId = (.*)$`)
    for snl.Scan() {
        currText := snl.Text()
        if reCluster.MatchString(currText) {
            clusterId = reCluster.ReplaceAllString(currText, `$1`)
        }
        if reProc.MatchString(currText) {
            procId = reProc.ReplaceAllString(currText, `$1`)
        }
    }
    err = snl.Err()
    if err != nil {
        fmt.Println("File reading error")
        return "ERROR"
    }
    return fmt.Sprintf("%s.%s", clusterId, procId)
}
