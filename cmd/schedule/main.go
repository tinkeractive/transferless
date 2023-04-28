package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/rclone/rclone/backend/s3"
	"github.com/tinkeractive/transferless/pkg/configuration"
	"github.com/tinkeractive/transferless/pkg/enqueuer"
	"github.com/tinkeractive/transferless/pkg/scheduler"
)

func main() {
	if "" == os.Getenv("AWS_LAMBDA_FUNCTION_NAME") {
		HandleRequest()
	} else {
		lambda.Start(HandleRequest)
	}
}

func HandleRequest() {
	region := os.Getenv("AWS_REGION")
	remote := os.Getenv("TRANSFERLESS_JOB_CONFIG_REMOTE")
	objPath := os.Getenv("TRANSFERLESS_JOB_CONFIG_PATH")
	jobQueue := os.Getenv("TRANSFERLESS_JOB_QUEUE")
	remoteConfigService := os.Getenv("TRANSFERLESS_REMOTE_CONFIG_SERVICE")
	var configProvider interface{}
	switch remoteConfigService {
	case "AWSSecretsManager":
		configProvider = &configuration.AWSSecretsManager{"Type", "Transferless"}
	case "AWSSystemsManager":
		configProvider = &configuration.AWSSystemsManager{"Type", "Transferless"}
	default:
		log.Println("no credential service specified")
		return
	}
	configString, err := configProvider.(configuration.Provider).GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("loading config")
	log.Println(configString)
	err = configuration.LoadConfig(context.Background(), configString)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("getting jobs config")
	jobs, err := scheduler.GetJobs(remote, objPath)
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
