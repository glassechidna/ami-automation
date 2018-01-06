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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"time"
)

var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Waits for an AMI (or multiple) to be available",
	Run: func(cmd *cobra.Command, args []string) {
		amiIds, _ := cmd.PersistentFlags().GetStringSlice("image-id")
		regions, _ := cmd.PersistentFlags().GetStringSlice("region")

		if len(amiIds) != len(regions) {
			fmt.Fprintf(os.Stderr, "The number of AMIs (%d) must match the number of regions (%d)\n", len(amiIds), len(regions))
		}

		regionalAmis := map[string]string{}
		for idx := range amiIds {
			regionalAmis[regions[idx]] = amiIds[idx]
		}

		wait(awsSession(), regionalAmis)
	},
}

func wait(sess *session.Session, amiIds map[string]string) {
	for region, amiId := range amiIds {
		regionSess := sess.Copy(&aws.Config{Region: &region})
		api := ec2.New(regionSess)

		for {
			resp, err := api.DescribeImages(&ec2.DescribeImagesInput{
				ImageIds: []*string{ &amiId },
			})

			if err != nil { panic(err) }
			if *resp.Images[0].State == ec2.ImageStateAvailable { break }

			time.Sleep(5 * time.Second)
		}
	}
}

func init() {
	utilCmd.AddCommand(waitCmd)
	waitCmd.PersistentFlags().StringSliceP("image-id", "i", []string{""}, "(Multiple) AMI IDs")
	waitCmd.PersistentFlags().StringSliceP("region", "r", []string{""}, "(Multiple) Regions hosting AMI IDs (in same order)")
}
