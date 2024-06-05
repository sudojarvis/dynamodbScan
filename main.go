package main

import (
	"context"
	"fmt"
	"log"
)

func main() {
	roleArn := "arn:aws:iam::{account-id}:role/{role-name}"
	roleSessionName := "demo-sessions"

	accessKey := "ACCESS_KEY"
	secretKey := "SECRET_KEY"
	creds, err := assumeRole(accessKey, secretKey, roleArn, roleSessionName)
	if err != nil {
		log.Fatalf("Error assuming role: %v", err)
	}
	fmt.Println("Temporary credentials:", creds)

	callerIdentity, err := GetCallerIdentity(creds)
	if err != nil {
		log.Fatalf("Error getting caller identity: %v", err)
	}

	fmt.Printf("User Account: %s\n", *callerIdentity.Account)
	fmt.Printf("User ARN: %s\n", *callerIdentity.Arn)
	fmt.Printf("User ID: %s\n", *callerIdentity.UserId)

	sourceTable := "test_tb"
	destinationTable := "test_tb_copy"
	totalSegments := 10
	region := "us-east-1"

	parallelCopy(context.Background(), creds, sourceTable, destinationTable, region, totalSegments)

	// parallelCopy(creds, sourceTable, destinationTable, totalSegments, region, 10)

	err = LaunchEC2Instance(creds, "us-east-1")
	if err != nil {
		log.Fatalf("Error launching instance: %v", err)
	}
}
