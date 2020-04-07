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
    db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
    if err != nil {
        fmt.Printf("Error while creating DB connection (receiver)")
        return err
    }
    defer db.Close()
    stmt, err := db.Prepare(`UPDATE htcjobs SET status=4 WHERE htcjobName=$1 AND jobid=$2`)
    if err != nil {
        fmt.Println("Error while creating DB statement (receiver)")
        return err
    }
    _, err = stmt.Exec(htcjobName, jobId)
    if err != nil{
        fmt.Println("Error while executing DB statement (receiver)")
        return err
    }
    return nil
}
