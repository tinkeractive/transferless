package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tinkeractive/transferless/configuration"
	"github.com/tinkeractive/transferless/enqueuer"
	"github.com/tinkeractive/transferless/scheduler"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/rclone/rclone/backend/s3"
)

func main(){
	if ("" == os.Getenv("AWS_LAMBDA_FUNCTION_NAME")) {
		HandleRequest()
	} else {
		lambda.Start(HandleRequest)		
	}
}

func HandleRequest() {
	region := os.Getenv("AWS_REGION")
	remote := os.Getenv("TRANSFERLESS_JOB_CONFIG_REMOTE")
	bucket := os.Getenv("TRANSFERLESS_JOB_CONFIG_BUCKET")
	objPath := os.Getenv("TRANSFERLESS_JOB_CONFIG_PATH")
	jobQueue := os.Getenv("TRANSFERLESS_JOB_QUEUE")
	configProvider := configuration.AWS{"Type", "Transferless"}
	configString, err := configProvider.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("loading config")
	err = configuration.LoadConfig(context.Background(), configString)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("getting jobs config")
	sourcePath := fmt.Sprintf("%s/%s", bucket, objPath)
	jobs, err := scheduler.GetJobs(remote, sourcePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("creating jobs enqueuer")
	awsEnqueuer, err := enqueuer.NewAWSEnqueuer(region, jobQueue, "")
	if err != nil {
		log.Fatal(err)
	}
	for _, job := range jobs {
		log.Println("enqueueing:", job)
		err = awsEnqueuer.EnqueueJob(job)
		if err != nil {
			log.Println(err)
		}
	}
}