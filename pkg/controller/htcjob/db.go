package htcjob

import (
    "database/sql"
    "fmt"
    "bufio"
    "os"
    "regexp"
    "path"

    _ "github.com/mattn/go-sqlite3"
)

func recordSubmission(htcjobName string, tempDirName string) ([]string, error) {
    // get job id
    var clusterId, procId []string

    buf, err := os.Open(path.Join(tempDirName, "condor_output.txt"))
    if err != nil {
        fmt.Println("File reading error")
        return nil, err
    }
    defer buf.Close()

    snl := bufio.NewScanner(buf)
    reCluster := regexp.MustCompile(`^ClusterId = (.*)$`)
    reProc := regexp.MustCompile(`^ProcId = (.*)$`)
    for snl.Scan() {
        currText := snl.Text()
        if reCluster.MatchString(currText) {
            clusterId = append(clusterId, reCluster.ReplaceAllString(currText, `$1`))
        }
        if reProc.MatchString(currText) {
            procId = append(procId, reProc.ReplaceAllString(currText, `$1`))
        }
    }
    err = snl.Err()
    if err != nil {
        fmt.Println("File reading error")
        return nil, err
    }
    var jobId []string
    for i := range clusterId {
        currentJobId := fmt.Sprintf("%s.%s", clusterId[i], procId[i])
        jobId = append(jobId, currentJobId)
        // inset record into the DB
        db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
        if err != nil {
            fmt.Printf("Error while inserting the job into DB")
            return nil, err
        }
        defer db.Close()
        stmt, err := db.Prepare(`INSERT INTO htcjobs VALUES (?, ?, ?, ?)`)
        if err != nil {
            fmt.Printf("Error while preparing a DB statement")
            return nil, err
        }
        _, err = stmt.Exec(htcjobName, currentJobId, 1, tempDirName)
        if err != nil {
            fmt.Printf("Error while inserting with a DB statement")
            return nil, err
        }
    }
    return jobId, nil
}

func getJobStatus(htcjobName string, jobId string) (int, error) {
    var status int
    db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
    if err != nil {
        return 0, err
    }
    defer db.Close()
    rows, err := db.Query(`SELECT status FROM htcjobs WHERE htcjobName=? AND jobId=?`,
        htcjobName, jobId)
    if err != nil {
        fmt.Printf("Error while preparing a DB statement")
        return 0, err
    }
    rows.Next()
    err = rows.Scan(&status)
    if err != nil {
        return 0, err
    }
    rows.Close()
    return status, nil
}
