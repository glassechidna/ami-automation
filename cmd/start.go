// Copyright © 2017 Aidan Steele <aidan.steele@glassechidna.com.au>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/aws/aws-sdk-go/service/ssm"
	"strings"
	"github.com/glassechidna/ami-automation/shared"
	"os"
	"github.com/fatih/color"
	"encoding/json"
	"fmt"
	"log"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an SSM automation execution",
	Run: func(cmd *cobra.Command, args []string) {
		name := viper.GetString("name")
		version:= viper.GetString("version")

		rawParameters := viper.GetStringSlice("parameter")
		params := parseRawParameters(rawParameters)

		accounts := viper.GetStringSlice("account")
		regions := viper.GetStringSlice("region")

		shouldWait := viper.GetBool("copy-wait")

		if len(accounts) > 0 && !shouldWait {
			fmt.Fprintln(os.Stderr, "You must wait (-w) if you want to share AMIs with other accounts. See GitHub issue #1.")
			os.Exit(1)
		}

		execId, err := start(name, version, params, accounts, regions)
		if err != nil { log.Panic(err.Error()) }

		sess := awsSession()
		reporter := shared.NewStatusReporter(sess, execId)
		reporter.Print()

		if !reporter.Success() {
			os.Exit(1)
		}

		amiId := reporter.AmiIds()[0]
		regionalAmis := copyAmiUi(sess, amiId, regions)

		if shouldWait {
			color.New(color.FgBlue).Fprintln(os.Stderr, "Waiting for copied AMIs to be available")
			wait(sess, regionalAmis)
			shareAmiUi(sess, regionalAmis, accounts)
		}

		output := shared.OutputFormat{
			Outputs: reporter.Outputs(),
			AmiId: amiId,
			AmiIds: regionalAmis,
			WaitCommand: makeWaitCommand(regionalAmis),
		}

		outputBytes, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(outputBytes))
	},
}

func makeWaitCommand(amiIds map[string]string) string {
	cmd := fmt.Sprintf("%s util wait", os.Args[0])

	for region, amiId := range amiIds {
		cmd = fmt.Sprintf("%s -i %s -r %s", cmd, amiId, region)
	}

	return cmd
}

func parseRawParameters(rawParameters []string) map[string][]*string {
	params := map[string][]*string{}
	for _, raw := range rawParameters {
		pair := strings.SplitN(raw, "=", 2)
		key := pair[0]
		val := pair[1]

		ary := params[key]
		if ary == nil {
			ary = []*string{}
		}

		params[key] = append(ary, &val)
	}

	return params
}

func start(name, version string, parameters map[string][]*string, accounts, regions []string) (string, error) {
	sess := awsSession()
	api := ssm.New(sess)

	input := &ssm.StartAutomationExecutionInput{
		DocumentName:    &name,
		Parameters:      parameters,
	}

	if len(version) > 0 {
		input.DocumentVersion = &version
	}

	resp, err := api.StartAutomationExecution(input)
	if err != nil { return "", err }

	execId := *resp.AutomationExecutionId
	return execId, nil
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("name", "", "SSM Automation document name")
	startCmd.PersistentFlags().String("version", "", "(optional) document version")
	startCmd.PersistentFlags().StringSliceP("parameter", "p", []string{""}, "(optional, multiple) document input parameters")
	startCmd.PersistentFlags().StringSliceP("region", "r", []string{""}, "(optional, multiple) AWS regions to copy AMI to")
	startCmd.PersistentFlags().StringSliceP("account", "a", []string{""}, "(optional, multiple) AWS accounts to share AMI with")
	startCmd.PersistentFlags().BoolP("copy-wait", "w", false, "Wait for copied images to be available")

	viper.BindPFlags(startCmd.PersistentFlags())
}
