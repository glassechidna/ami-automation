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
	"log"
	"github.com/glassechidna/ami-automation/shared"
	"os"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an SSM automation execution",
	Run: func(cmd *cobra.Command, args []string) {
		name := viper.GetString("name")
		version:= viper.GetString("version")

		rawParameters := viper.GetStringSlice("parameter")
		params := parseRawParameters(rawParameters)

		execId, err := start(name, version, params)
		if err != nil { log.Panic(err.Error()) }

		sess := awsSession()
		reporter := shared.NewStatusReporter(sess, execId)
		reporter.Print()

		if reporter.Success() {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	},
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

func start(name, version string, parameters map[string][]*string) (string, error) {
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

	viper.BindPFlags(startCmd.PersistentFlags())
}
