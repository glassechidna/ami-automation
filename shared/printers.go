package shared

import (
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"github.com/fatih/color"
	"strings"
	"os"
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"encoding/json"
	"strconv"
)

type StepPrinter interface {
	Print(file *os.File, sess *session.Session, step *ssm.StepExecution) error
}

type DefaultPrinter struct {}

func (p *DefaultPrinter) Print(file *os.File, sess *session.Session, step *ssm.StepExecution) error {
	fmt.Fprintf(file, "Unhandled step action: %s\n", *step.Action)
	return nil
}

type RunInstancesPrinter struct {}

func (p *RunInstancesPrinter) Print(file *os.File, sess *session.Session, step *ssm.StepExecution) error {
	idPtrs := step.Outputs["InstanceIds"]
	ids := []string{}

	for _, idptr := range idPtrs {
		ids = append(ids, *idptr)
	}

	joined := strings.Join(ids, ", ")
	fmt.Fprintf(file, "Instance IDs: %s\n", joined)
	return nil
}

type InvokeLambdaPrinter struct {}

func (p *InvokeLambdaPrinter) Print(file *os.File, sess *session.Session, step *ssm.StepExecution) error {
	rawInput := *step.Inputs["Payload"]
	rawOutput := *step.Outputs["Payload"][0]
	unquotedInput, _ := strconv.Unquote(rawInput)

	input := prettyPrintedMaybeJson(unquotedInput)
	output := prettyPrintedMaybeJson(rawOutput)

	fmt.Fprintf(file, "Input: %s\nOutput: %s\n", color.GreenString(input), color.GreenString(output))
	return nil
}

func prettyPrintedMaybeJson(input string) string {
	msg := json.RawMessage{}

	err := json.Unmarshal([]byte(input), &msg)
	if err != nil { return input }

	bytes, err := json.MarshalIndent(msg, "", "  ")
	if err != nil { return input }

	return string(bytes)
}

type RunCommandPrinter struct {}

func (p *RunCommandPrinter) Print(file *os.File, sess *session.Session, step *ssm.StepExecution) error {
	api := s3.New(sess)

	color.New(color.FgBlue, color.Bold).Fprint(file, *step.StepName)
	color.New(color.FgBlue).Fprintf(file, ": %s\n", *step.StepStatus)

	bucket := step.Inputs["OutputS3BucketName"]

	if bucket != nil {
		bucketName, err := strconv.Unquote(*bucket)
		if err != nil { return err }

		commandId := step.Outputs["CommandId"]
		keyPrefix := *commandId[0]
		keyPrefixPtr := step.Inputs["OutputS3KeyPrefix"]
		if keyPrefixPtr != nil {
			keyPrefix = fmt.Sprintf("%s/%s", *keyPrefixPtr, keyPrefix)
		}

		listResp, err := api.ListObjects(&s3.ListObjectsInput{
			Bucket: &bucketName,
			Prefix: &keyPrefix,
		})
		if err != nil { return err }

		for _, object := range listResp.Contents {
			getResp, err := api.GetObject(&s3.GetObjectInput{
				Bucket: &bucketName,
				Key: object.Key,
			})
			if err != nil { return err }

			body, err := ioutil.ReadAll(getResp.Body)
			if err != nil { return err }

			bodyColor := color.New(color.FgGreen)
			if strings.HasSuffix(*object.Key, "stderr") {
				bodyColor = color.New(color.FgRed)
			}

			bodyColor.Fprintln(file, string(body))
		}
	}

	return nil
}

type CreateImagePrinter struct {}

func (p *CreateImagePrinter) Print(file *os.File, sess *session.Session, step *ssm.StepExecution) error {
	idPtrs := step.Outputs["ImageId"]
	ids := []string{}

	for _, idptr := range idPtrs {
		ids = append(ids, *idptr)
	}

	joined := strings.Join(ids, ", ")
	fmt.Fprintf(file, "Image ID: %s\n", joined)
	return nil
}
