package htcjob

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

func recordSubmission(htcjobName string, tempDirName string, uniqid int) ([]string, error) {
	// get job id
	var clusterID, procID []string

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
			clusterID = append(clusterID, reCluster.ReplaceAllString(currText, `$1`))
		}
		if reProc.MatchString(currText) {
			procID = append(procID, reProc.ReplaceAllString(currText, `$1`))
		}
	}
	err = snl.Err()
	if err != nil {
		fmt.Println("File reading error")
		return nil, err
	}
	var jobID []string
	for i := range clusterID {
		currentJobID := fmt.Sprintf("%s.%s", clusterID[i], procID[i])
		jobID = append(jobID, currentJobID)
		// inset record into the DB
		db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
		if err != nil {
			fmt.Printf("Error while inserting the job into DB")
			return nil, err
		}
		defer db.Close()
		stmt, err := db.Prepare(`INSERT INTO htcjobs VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			fmt.Printf("Error while preparing a DB statement")
			return nil, err
		}
		_, err = stmt.Exec(htcjobName, currentJobID, 1, tempDirName, uniqid)
		if err != nil {
			fmt.Printf("Error while inserting with a DB statement")
			return nil, err
		}
	}
	return jobID, nil
}

func updateJobStatus(htcjobName string, jobID string, newStatus int) error {

	db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
	if err != nil {
		return fmt.Errorf("Error while inserting the job into DB: %v", err)
	}
	defer db.Close()

	fmt.Printf("updateJobStatus for htcjobName: %s, jobID: %s, newStatus: %d\n", htcjobName, jobID, newStatus)
	stmt, err := db.Prepare(`UPDATE htcjobs SET status=$1 WHERE htcjobName=$2 AND jobId=$3`)
	if err != nil {
		return fmt.Errorf("Error while creating DB statement (updateJobStatus): %v", err)
	}
	_, err = stmt.Exec(newStatus, htcjobName, jobID)
	if err != nil {
		return fmt.Errorf("Error while executing DB statement (updateJobStatus): %v", err)
	}
	return nil
}

func getJobStatus(htcjobName string, jobID string) (int, error) {
	var status int
	db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
	if err != nil {
		return 0, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT status FROM htcjobs WHERE htcjobName=? AND jobId=?`,
		htcjobName, jobID)
	if err != nil {
		fmt.Printf("Error while preparing a DB statement")
		return 0, err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&status)
	if err != nil {
		return 0, err
	}
	return status, nil
}

func getJobTempDir(htcjobName string, jobID string) (string, error) {
	tempDirName := ""
	db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
	if err != nil {
		return tempDirName, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT tempDir FROM htcjobs WHERE htcjobName=? AND jobId=?`,
		htcjobName, jobID)
	if err != nil {
		fmt.Printf("Error while preparing a DB statement")
		return "", err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&tempDirName)
	if err != nil {
		return tempDirName, err
	}
	return tempDirName, nil
}

func rmJob(htcjobName string, jobID string) error {
	db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
	if err != nil {
		fmt.Printf("Error while deleting the job from the DB (connection)")
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare(`DELETE FROM htcjobs WHERE jobid = ? AND htcjobname = ?`)
	if err != nil {
		fmt.Printf("Error while preparing a DB statement while removing")
		return err
	}
	_, err = stmt.Exec(jobID, htcjobName)
	if err != nil {
		fmt.Printf("Error while deleting from DB")
		return err
	}
	return nil
}

func clearJobs(htcjobName string, uniqID int) error {
	db, err := sql.Open("sqlite3", "/data/sqlite/htcjobs.db")
	if err != nil {
		fmt.Printf("Error while deleting the job from the DB (connection) (clear)")
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare(`DELETE FROM htcjobs WHERE uniqid = ? AND htcjobname = ?`)
	if err != nil {
		fmt.Printf("Error while preparing a DB statement while removing (clear)")
		return err
	}
	_, err = stmt.Exec(uniqID, htcjobName)
	if err != nil {
		fmt.Printf("Error while deleting from DB( clear)")
		return err
	}
	return nil
}
