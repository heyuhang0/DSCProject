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

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "For Client to send a GET request to Server",
	Long: `Description: For Client to send a GET request to Server.
			Input Argument: 
				1. Address of Server (type string) i.e localhost:50051,
				2. Key (type string)
			Expected Response:
				GET SUCCESSFUL, {key, value}`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== GET Request is called! ===")

		address, key := args[0], args[1]

		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Did not connect: %v", err)
		}
		defer func() { _ = conn.Close() }()
		c := pb.NewKeyValueStoreClient(conn)

		// Basic Test
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		resp, err := c.Get(ctx, &pb.GetRequest{Key: []byte(key)})
		if err != nil {
			log.Fatalf("Failed to GET: %v", err)
		}
		if resp.FoundKey == pb.FoundKey_KEY_NOT_FOUND{
			log.Printf("GET SUCCESSFUL: key not found")
		}else{
			log.Printf("GET SUCCESSFUL: %v %v", key,string(resp.Object))
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
