/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"context"
	"fmt"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"google.golang.org/grpc"
	"log"
	"time"

	"github.com/spf13/cobra"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "For Client to send a PUT request to Server",
	Long: `Description: For Client to send a PUT request to Server.
			Input Argument: 
				1. Address of Server (type string) i.e localhost:50051,
				2. Key (type string)
				3. Value (type string)
			Output: 
				returns {key: value} inserted`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== PUT Request is called! ===")

		address, key, value := args[0], args[1], args[2]

		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Did not connect: %v", err)
		}
		defer func() { _ = conn.Close() }()
		c := pb.NewKeyValueStoreClient(conn)

		// Basic Test
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = c.Put(ctx, &pb.PutRequest{
			Key:    []byte(key),
			Object: []byte(value),
		})
		if err != nil {
			log.Fatalf("Failed to PUT: %v", err)
		}

		log.Printf("PUT SUCCESSFUL: {%v: %v}", key, value)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// putCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// putCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
