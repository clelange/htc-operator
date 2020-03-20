package main

import (
    "fmt"
    "os"
    "log"
    "os/exec"
    "io/ioutil"
    "bufio"
    "regexp"
    "strings"
    "time"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
    accessKey := "PJE22MGQ8QO45CKI496K"
    secretKey := "8BnRJcxuOain8fLBxT44IfFMNgku7huHoo8mK8NP"
    s3Config := &aws.Config{
        Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
        Endpoint:         aws.String("s3.cern.ch"),
        Region:           aws.String("us-east-1"),
        DisableSSL:       aws.Bool(true),
        S3ForcePathStyle: aws.Bool(true),
    }
    bucket := "TADO_BUCKET"
    itemsS3 := s3Objects(s3Config, bucket)
    // job id and filename files
    tmpFile, err := ioutil.TempFile(os.TempDir(), "")
    if err != nil {
        fmt.Println("Cannot create temporary file", err)
        return
    }
    defer os.Remove(tmpFile.Name())
    idsFiles := make([]string, 0)
    for _, i := range itemsS3 {
        if matched, _ := regexp.MatchString(`run_`, i); matched {
            idList := getJobIds(s3Config, bucket, i)
            for _, id := range idList {
                idsFiles = append(idsFiles, id + " " + i)
            }
        }
    }
    // don't proceed if no IDs found
    if len(idsFiles) == 0 {
        return
    }
    // write info to file for python to read
    writer := bufio.NewWriter(tmpFile)
    for _, l := range idsFiles {
        writer.WriteString(l +"\n")
    }
    writer.Flush()
    // call python script
    fmt.Println("HTC query @ " + time.Now().String())
    out, err := exec.Command("python", "/scratch/runCondorQ.py", tmpFile.Name()).Output()
    if err != nil {
        fmt.Printf("Failed to run the pyhon script: %s", err)
        return
    }
    lines := strings.Split(string(out), "\n")
    lines = lines[:(len(lines) - 1)]
    pendingJobs := make(map[string]int)
    for _, currLine := range lines {
        splitLine := strings.Split(currLine, " ")
        pendingJobs[splitLine[1]] += 1
        if splitLine[2] == "4" {
            pendingJobs[splitLine[1]] -= 1
        }
    }
    reRep := regexp.MustCompile("run_")
    // move all with all completed
    for k, v := range pendingJobs {
        if v == 0 {
            moveS3Object(s3Config, bucket, k, reRep.ReplaceAllString(k, "complete_"))
        }
    }
}

func putS3Object(s3Config *aws.Config, bucket string, filename string, fileKey string) {
    svc := s3.New(session.New(s3Config))
    input := &s3.PutObjectInput{
        Body:                 aws.ReadSeekCloser(strings.NewReader(filename)),
        Bucket:               aws.String(bucket),
        Key:                  aws.String(fileKey),
    }
    result, err := svc.PutObject(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return
    }
    fmt.Println(result)
}

func deleteS3Object(s3Config *aws.Config, bucket string, fileKey string) {
    svc := s3.New(session.New(s3Config))
    input := &s3.DeleteObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(fileKey),
    }
    result, err := svc.DeleteObject(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return
    }
    fmt.Println(result)
}

func moveS3Object(s3Config *aws.Config, bucket string, nameA string, nameB string) {
    // save the file locally
    downloader := s3manager.NewDownloader(session.New(s3Config))
    input := &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(nameA),
    }
    // tempfile to store the downloaded file
    tmpFile, err := ioutil.TempFile(os.TempDir(), "")
    if err != nil {
        fmt.Println("Cannot create temporary file", err)
        return
    }
    defer os.Remove(tmpFile.Name())
    _, err = downloader.Download(tmpFile, input)
    if err != nil {
        fmt.Printf("Unable to download item %q, %v", nameA, tmpFile)
        return
    }
    // delete the old one
    deleteS3Object(s3Config, bucket, nameA)
    // add the new one
    putS3Object(s3Config, bucket, tmpFile.Name(), nameB)
}

func getJobIds(s3Config *aws.Config, bucket string, filename string) []string {
    downloader := s3manager.NewDownloader(session.New(s3Config))
    input := &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(filename),
    }
    // tempfile to store the downloaded file
    tmpFile, err := ioutil.TempFile(os.TempDir(), "")
    if err != nil {
        fmt.Println("Cannot create temporary file", err)
        return make([]string, 0)
    }
    defer os.Remove(tmpFile.Name())
    _, err = downloader.Download(tmpFile, input)
    if err != nil {
        log.Fatalf("Unable to download item %q, %v", filename, tmpFile)
    }

    // read the file
    file, err := os.Open(tmpFile.Name())
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    fscanner := bufio.NewScanner(file)
    resultIds := make([]string, 0)
    reRepId := regexp.MustCompile("^.*Proc (.*):$")
    for fscanner.Scan() {
        currentLine := fscanner.Text()
        if matched, _ := regexp.MatchString(`\*\* Proc`, currentLine); matched {
            resultIds = append(resultIds, reRepId.ReplaceAllString(currentLine, "$1"))
        }
    }
    return resultIds
}

func s3Objects(s3Config *aws.Config, bucket string) []string {
    svc := s3.New(session.New(s3Config))
    input := &s3.ListObjectsV2Input{
        Bucket:  aws.String(bucket),
        MaxKeys: aws.Int64(10000),
    }
    result, err := svc.ListObjectsV2(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            case s3.ErrCodeNoSuchBucket:
                fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return make([]string, 0)
    }
    itemSlice := make([]string, 0)
    for _, item := range result.Contents {
        itemSlice = append(itemSlice, *(item.Key))
    }
    return itemSlice
}
