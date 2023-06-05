package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	_ "github.com/rclone/rclone/backend/s3"
	_ "github.com/rclone/rclone/backend/sftp"
	"github.com/tinkeractive/transferless/pkg/compiler"
	"github.com/tinkeractive/transferless/pkg/configuration"
	"github.com/tinkeractive/transferless/pkg/enqueuer"
	"github.com/tinkeractive/transferless/pkg/job"
	"github.com/tinkeractive/transferless/pkg/transfer"
)

// NOTE the on-disk config requirement is unavoidable without altering the rclone source code
// NOTE the rclone config functionality consists of types with unexported attributes and methods
// NOTE the config can contain fields that are not required by rclone and they will be parsed
// NOTE rclone obscures passwords before saving in the config file and users must do the same
// NOTE the aws config provider assumes secrets manager usage

func main() {
	if "" == os.Getenv("AWS_LAMBDA_FUNCTION_NAME") {
		region := os.Getenv("AWS_REGION")
		// NOTE this env var does not exist in deployment config
		jobQueue := os.Getenv("TRANSFERLESS_JOB_QUEUE")
		sqsClient := sqs.New(session.New(), &aws.Config{
			Region: aws.String(region),
		})
		getQueueURLInput := &sqs.GetQueueUrlInput{
			QueueName: aws.String(jobQueue),
		}
		getQueueURLOutput, err := sqsClient.GetQueueUrl(getQueueURLInput)
		if err != nil {
			log.Fatal(err)
		}
		receiveMessageInput := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(*getQueueURLOutput.QueueUrl),
			MaxNumberOfMessages: aws.Int64(int64(1)),
		}
		receiveMessageOutput, err := sqsClient.ReceiveMessage(receiveMessageInput)
		if err != nil {
			log.Fatal(err)
		}
		var inputJob job.Job
		err = json.Unmarshal([]byte(*receiveMessageOutput.Messages[0].Body), &inputJob)
		if err != nil {
			log.Fatal(err)
		}
		CompileJob(inputJob)
	} else {
		lambda.Start(HandleRequest)
	}
}

func HandleRequest(lambdaEvent events.SQSEvent) {
	for _, event := range lambdaEvent.Records {
		var inputJob job.Job
		err := json.Unmarshal([]byte(event.Body), &inputJob)
		if err != nil {
			log.Fatal(err)
		}
		CompileJob(inputJob)
	}
}

func CompileJob(inputJob job.Job) {
	log.Println("job:", inputJob)
	region := os.Getenv("AWS_REGION")
	rawRemote := os.Getenv("TRANSFERLESS_DATA_REMOTE")
	remote := strings.ReplaceAll(rawRemote, "/", "")
	dataRoot := os.Getenv("TRANSFERLESS_DATA_ROOT")
	transferQueue := os.Getenv("TRANSFERLESS_TRANSFER_QUEUE")
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
	err = configuration.LoadConfig(context.Background(), configString)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("checking job lock")
	isLocked, err := compiler.IsLocked(remote, dataRoot, inputJob.Name)
	if err != nil {
		log.Fatal(err)
	}
	if isLocked {
		log.Println(inputJob.Name, "is locked")
		return
	}
	log.Println("locking job")
	err = compiler.Lock(remote, dataRoot, inputJob.Name)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("getting last mod time")
	lastModTime, err := compiler.GetLastModTime(remote, dataRoot, inputJob.Name)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("compiling job transfers")
	transfers, err := compiler.Compile(inputJob, lastModTime)
	if err != nil {
		log.Fatal(err)
	}
	maxModTime := lastModTime
	awsEnqueuer, err := enqueuer.NewAWSEnqueuer(region, "", transferQueue)
	if err != nil {
		log.Fatal(err)
	}
	_ = awsEnqueuer
	for _, transferFile := range transfers {
		log.Println("enqueueing:", transferFile)
		err = awsEnqueuer.EnqueueTransfer(transfer.Transfer{transferFile, inputJob})
		if err != nil {
			log.Println("compiler failed to enqueue:", inputJob, transferFile)
		}
		if err == nil && maxModTime < transferFile.LastModified {
			maxModTime = transferFile.LastModified
		}
	}
	log.Println("putting max mod time", maxModTime)
	err = compiler.PutModTime(remote, dataRoot, inputJob.Name, maxModTime)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("unlocking job")
	err = compiler.Unlock(remote, dataRoot, inputJob.Name)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("exiting")
}
