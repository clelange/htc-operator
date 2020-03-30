package htcjob

import (
    "database/sql"
    "fmt"
    "bufio"
    "os"
    "regexp"
    "path"

    _ "github.com/lib/pq"
)

const (
    host         = "cms-batch-test.cern.ch"
    port         = 30303
    user         = "postgres"
    password     = "pgpasswd"
    dbname       = "postgres"
)

func recordSubmission(htcjobName string, tempDirName string) (string, error) {
    // get job id
    var clusterId, procId string

    buf, err := os.Open(path.Join(tempDirName, "condor_output.txt"))
    if err != nil {
        fmt.Println("File reading error")
        return "", err
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
        return "", err
    }
    jobId := fmt.Sprintf("%s.%s", clusterId, procId)
    // inset record into the DB
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        fmt.Printf("Error while inserting the job into DB")
        return "", err
    }
    defer db.Close()
    sqlStatement := `INSERT INTO htcjobs VALUES ($1, $2, $3, $4)`
    _, err = db.Exec(sqlStatement, htcjobName, jobId, 1, tempDirName)
    if err != nil {
        fmt.Printf("Error while inserting the job into DB")
        return "", err
    }
    return jobId, nil
}

func getJobStatus(htcjobName string, jobId string) (byte, error) {
    var status byte
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return 0, err
    }
    defer db.Close()
    sqlStatement := `
    SELECT status FROM htcjobs WHERE htcjobName=$1 AND jobId=$2`
    row := db.QueryRow(sqlStatement, htcjobName, jobId)
    err = row.Scan(&status)
    if err != nil {
        return 0, err
    }
    return status, nil
}
