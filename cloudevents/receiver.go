package main

import (
    "context"
    "fmt"
    "log"
    cloudevents "github.com/cloudevents/sdk-go/v2"
    "database/sql"
    "encoding/json"

    _ "github.com/mattn/go-sqlite3"
)

type Response struct {
    Name string `json:"name"`
    JobId string `json:"jobid"`
    RetCode string `json:"retcode"`
}

func main() {
    ctx := context.Background()
    p, err := cloudevents.NewHTTP(cloudevents.WithPort(8080))
    if err != nil {
        log.Fatalf("failed to create protocol: %s", err.Error())
    }

    c, err := cloudevents.NewClient(p)
    if err != nil {
        log.Fatalf("failed to create client, %v", err)
    }

    log.Printf("Listening on :80\n")
    log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, receive))
}

func receive(ctx context.Context, event cloudevents.Event) {
    resp := Response{}
    if err := json.Unmarshal(event.Data(), &resp); err != nil {
        panic(err)
    }
    if err := markJobDone(resp.Name, resp.JobId, resp.RetCode); err != nil {
        panic(err)
    }
}

func markJobDone(htcjobName string, jobId string, retCode string) error {
    db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
    if err != nil {
        fmt.Printf("Error while creating DB connection (receiver)")
        return err
    }
    defer db.Close()
    jobStatus := 4
    if retCode != "0" {
        jobStatus = 7 // 7 means 'error' here (not suspended)
    }
    fmt.Printf("jobname: %s, jobid: %s, retcode: %s\n", htcjobName, jobId, retCode)
    stmt, err := db.Prepare(`UPDATE htcjobs SET status=$1 WHERE htcjobName=$2 AND jobid=$3`)
    if err != nil {
        fmt.Println("Error while creating DB statement (receiver)")
        return err
    }
    _, err = stmt.Exec(jobStatus, htcjobName, jobId)
    if err != nil{
        fmt.Println("Error while executing DB statement (receiver)")
        return err
    }
    return nil
}
