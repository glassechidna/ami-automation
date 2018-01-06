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

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		amiId, _ := cmd.PersistentFlags().GetString("image-id")
		regions, _ := cmd.PersistentFlags().GetStringSlice("region")
		shouldWait, _ := cmd.PersistentFlags().GetBool("wait")

		sess := awsSession()
		amiIds := copyAmiUi(sess, amiId, regions)

		if shouldWait {
			color.New(color.FgBlue).Fprintln(os.Stderr, "Waiting for copied AMIs to be available")
			wait(sess, amiIds)
		}
	},
}

type copyAmiResult struct {
	region string
	resp *ec2.CopyImageOutput
	err error
}

func copyAmiUi(sess *session.Session, amiId string, regions []string) map[string]string {
	regionalAmis := map[string]string{}

	blue := color.New(color.FgBlue)
	boldBlue := color.New(color.FgBlue, color.Bold)

	if len(regions) > 0 {
		boldBlue.Fprint(os.Stderr, "Copying AMI to other regions\n")
		regionalAmis = copyAmi(sess, amiId, regions)
		blue.Fprint(os.Stderr, "AMI IDs:\n")

		for region, amiId := range regionalAmis {
			blue.Fprintf(os.Stderr, "%s: %s\n", region, amiId)
		}
	}

	regionalAmis[*sess.Config.Region] = amiId
	return regionalAmis
}

func copyAmi(sess *session.Session, amiId string, regions []string) map[string]string {
	sourceRegion := *sess.Config.Region
	name := amiName(sess, amiId)
	amiIds := map[string]string{}

	for _, region := range regions {
		sess = sess.Copy(&aws.Config{Region: aws.String(region)})
		api := ec2.New(sess)

		resp, err := api.CopyImage(&ec2.CopyImageInput{
			SourceImageId: &amiId,
			SourceRegion: &sourceRegion,
			Name: &name,
		})
		if err != nil { panic(err) }

		amiIds[region] = *resp.ImageId
	}

	return amiIds
}

func init() {
	utilCmd.AddCommand(copyCmd)
	copyCmd.PersistentFlags().StringSliceP("region", "r", []string{""}, "(optional, multiple) AWS regions to copy AMI to")
	copyCmd.PersistentFlags().String("image-id", "", "Source AMI ID to copy")
	copyCmd.PersistentFlags().BoolP("wait", "w", false, "Wait for copied images to be available")
}
