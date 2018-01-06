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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"github.com/fatih/color"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		amiId, _ := cmd.PersistentFlags().GetString("image-id")
		accounts, _ := cmd.PersistentFlags().GetStringSlice("account")

		sess := awsSession()
		region := *sess.Config.Region
		regionalAmis := map[string]string{}
		regionalAmis[region] = amiId

		shareAmiUi(sess, regionalAmis, accounts)
	},
}

func shareAmiUi(sess *session.Session, regionalAmis map[string]string, accounts []string) {
	blue := color.New(color.FgBlue)
	boldBlue := color.New(color.FgBlue, color.Bold)

	if len(accounts) > 0 {
		boldBlue.Fprint(os.Stderr, "Sharing AMIs with other accounts\n")

		for _, amiId := range regionalAmis {
			shareAmi(sess, amiId, accounts)
			blue.Fprintf(os.Stderr, "Shared %s with %v\n", amiId, accounts)
		}
	}
}

func amiName(sess *session.Session, amiId string) string {
	api := ec2.New(sess)
	resp, _ := api.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{&amiId}})
	return *resp.Images[0].Name
}

func shareAmi(sess *session.Session, amiId string, accounts []string) {
	api := ec2.New(sess)

	permissions := []*ec2.LaunchPermission{}
	for _, account := range accounts {
		permissions = append(permissions, &ec2.LaunchPermission{
			UserId: aws.String(account),
		})
	}

	api.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
		ImageId: &amiId,
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: permissions,
		},
	})
}


func init() {
	utilCmd.AddCommand(shareCmd)
	shareCmd.PersistentFlags().String("image-id", "", "AMI ID to share")
	shareCmd.PersistentFlags().StringSliceP("account", "a", []string{""}, "(optional, multiple) AWS accounts to share AMI with")
}
