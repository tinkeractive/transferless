package enqueuer

import (
	"github.com/tinkeractive/transferless/pkg/job"
	"github.com/tinkeractive/transferless/pkg/transfer"
)

type Enqueuer interface {
	EnqueueJob(transferJob job.Job) error
	EnqueueTransfer(transferObj transfer.Transfer) error
}
