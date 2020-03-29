package main

import (
    "context"
    "log"
    "fmt"
    "os"
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
    fmt.Println(jobName)

    e := cloudevents.NewEvent()
    e.SetType("htcjob.cloudevent")
    e.SetSource("cern.ch")
    _ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
        "name": jobName,
    })

    err = c.Send(ctx, e)
    if err != nil {
        log.Printf("failed to send: %v", err)
    }
}

