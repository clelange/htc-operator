package main

import (
    "context"
    "fmt"
    "log"
    cloudevents "github.com/cloudevents/sdk-go/v2"
    "database/sql"
    "encoding/json"

    _ "github.com/lib/pq"
)

const (
    host         = "cms-batch-test.cern.ch"
    port         = 30303
    user         = "postgres"
    password     = "pgpasswd"
    dbname       = "postgres"
)

type Response struct {
    Name string `json:"name"`
    JobId string `json:"jobid"`
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
    if err := markJobDone(resp.Name, resp.JobId); err != nil {
        panic(err)
    }
}

func markJobDone(htcjobName string, jobId string) error {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        fmt.Printf("Error while inserting the job into DB")
        return err
    }
    defer db.Close()
    sqlStatement := `UPDATE htcjobs SET status=4 WHERE htcjobName=$1 AND jobid=$2`
    fmt.Println(sqlStatement)
    fmt.Println(htcjobName)
    fmt.Println(jobId)
    _, err = db.Exec(sqlStatement, htcjobName, jobId)
    if err != nil {
        fmt.Printf("Error while inserting the job into DB")
        return err
    }
    return nil
}
