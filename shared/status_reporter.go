package shared

import (
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
	"github.com/aws/aws-sdk-go/aws/session"
	"time"
	"github.com/fatih/color"
	"os"
)

func isTerminalStatus(status string) bool {
	switch status {
	case "Success", "TimedOut", "Cancelled", "Failed": return true
	default: return false
	}
}

func isSuccessStatus(status string) bool {
	return status == "Success"
}

func stringInSlice(str string, list []string) bool {
	for _, value := range list {
		if value == str {
			return true
		}
	}
	return false
}

type StatusReporter struct {
	sess *session.Session
	execId string
	progress *os.File
	results *os.File
}

func NewStatusReporter(sess *session.Session, execId string) *StatusReporter {
	return &StatusReporter{
		sess: sess,
		execId: execId,
		progress: os.Stderr,
		results: os.Stdout,
	}
}

func (r *StatusReporter) Success() bool {
	api := ssm.New(r.sess)
	resp, _ := api.GetAutomationExecution(&ssm.GetAutomationExecutionInput{
		AutomationExecutionId: &r.execId,
	})
	return *resp.AutomationExecution.AutomationExecutionStatus == "Success"
}

func (r *StatusReporter) Print() {
	color.New(color.FgBlue).Fprintf(r.progress, "SSM Automation execution ID: %s\n", r.execId)

	r.PrintSteps()
}

func (r *StatusReporter) Outputs() map[string][]*string {
	api := ssm.New(r.sess)
	resp, err := api.GetAutomationExecution(&ssm.GetAutomationExecutionInput{
		AutomationExecutionId: &r.execId,
	})
	if err != nil { log.Panicf(err.Error()) }

	return resp.AutomationExecution.Outputs
}

func (r *StatusReporter) PrintSteps() {
	api := ssm.New(r.sess)

	printedSteps := []string{}

	for {
		resp, err := api.GetAutomationExecution(&ssm.GetAutomationExecutionInput{
			AutomationExecutionId: &r.execId,
		})
		if err != nil { log.Panicf(err.Error()) }

		for _, step := range resp.AutomationExecution.StepExecutions {
			if isTerminalStatus(*step.StepStatus) && !stringInSlice(*step.StepName, printedSteps) {
				printedSteps = append(printedSteps, *step.StepName)
				r.PrintStep(step)
			}
		}

		if isTerminalStatus(*resp.AutomationExecution.AutomationExecutionStatus) {
			break
		}

		time.Sleep(5 * time.Second)
	}
}

func printerForType(stepType string) StepPrinter {
	switch stepType {
	case "aws:runCommand":
		return &RunCommandPrinter{}
	case "aws:invokeLambdaFunction":
		return &InvokeLambdaPrinter{}
	case "aws:runInstances":
		return &RunInstancesPrinter{}
	case "aws:createImage":
		return &CreateImagePrinter{}
	default:
		return &DefaultPrinter{}
	}
}

func (r *StatusReporter) PrintStep(step *ssm.StepExecution) error {
	color.New(color.FgBlue, color.Bold).Fprint(r.progress, *step.StepName)
	color.New(color.FgBlue).Fprintf(r.progress, ": %s\n", *step.StepStatus)

	printer := printerForType(*step.Action)
	return printer.Print(r.progress, r.sess, step)
}

func (r *StatusReporter) AmiIds() []string {
	api := ssm.New(r.sess)

	amiIds := []string{}

	resp, err := api.GetAutomationExecution(&ssm.GetAutomationExecutionInput{
		AutomationExecutionId: &r.execId,
	})
	if err != nil {
		log.Panicf(err.Error())
	}

	for _, step := range resp.AutomationExecution.StepExecutions {
		if *step.Action == "aws:createImage" {
			amiId := *step.Outputs["ImageId"][0]
			amiIds = append(amiIds, amiId)
		}
	}

	return amiIds
}
