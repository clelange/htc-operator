package main

import (
    "os"
    "io/ioutil"
    "bufio"
    "regexp"
    "strings"
    "context"
    "fmt"
    "log"
    "encoding/json"
    cloudevents "github.com/cloudevents/sdk-go/v2"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Response struct {
    Name string `json:"name"`
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
    jobName := resp.Name
    runName := fmt.Sprintf("run_%s_condor.out", jobName)
    completeName := fmt.Sprintf("complete_%s_condor.out", jobName)
    // download file
    downloader := s3manager.NewDownloader(session.New(s3Config))
    input := &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(runName),
    }
    // tempfile to store the downloaded file
    tmpFile, err := ioutil.TempFile(os.TempDir(), "")
    defer os.Remove(tmpFile.Name())
    if err != nil {
        fmt.Println("Cannot create temporary file", err)
        return
    }
    // download
    _, err = downloader.Download(tmpFile, input)
    if err != nil {
        fmt.Printf("Unable to download item %q, %v", runName, tmpFile)
        return
    }
    // write ids to fileB
    fileRead, err := os.Open(tmpFile.Name())
    if err != nil {
        log.Fatal(err)
    }
    reRepId := regexp.MustCompile("^.*Proc (.*):$")
    fscanner := bufio.NewScanner(tmpFile)
    var outputString string
    for fscanner.Scan() {
        currentLine := fscanner.Text()
        if matched, _ := regexp.MatchString(`\*\* Proc`, currentLine); matched {
            outputString += fmt.Sprintf("%s\n", reRepId.ReplaceAllString(currentLine, "$1"))
        }
    }
    fileRead.Close()
    deleteS3Object(s3Config, bucket, runName)
    // add the new one
    putS3Object(s3Config, bucket, outputString, completeName)
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
