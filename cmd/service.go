/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/allar/kitgen/kitgen"
	"log"

	"github.com/spf13/cobra"
)

var serviceName string

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Generates a service",
	Long:  `Generates a service`,
	Run: func(cmd *cobra.Command, args []string) {

		err := kitgen.CreateService(serviceName)
		if err != nil {
			log.Fatalln(fmt.Sprintf("Failed to create service. Error: %s", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	serviceCmd.Flags().StringVarP(&serviceName, "name", "n", "", "name of the new service")
	serviceCmd.MarkFlagRequired("name")
}
