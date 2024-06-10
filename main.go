package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func assumeRole(accessKey, secretKey, roleArn, roleSessionName string) (ststypes.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	if err != nil {
		return ststypes.Credentials{}, fmt.Errorf("unable to load SDK config: %v", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(roleSessionName),
	}

	result, err := stsClient.AssumeRole(context.TODO(), input)
	if err != nil {
		return ststypes.Credentials{}, fmt.Errorf("error assuming role: %v", err)
	}

	return *result.Credentials, nil
}

func CreateAWSConfigWithTempCredentials(creds *ststypes.Credentials, region string) (aws.Config, error) {
	// Create a custom credentials provider with temporary credentials
	credsProvider := credentials.NewStaticCredentialsProvider(
		aws.ToString(creds.AccessKeyId),
		aws.ToString(creds.SecretAccessKey),
		aws.ToString(creds.SessionToken),
	)

	// Load configuration with the temporary credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credsProvider),
		config.WithRegion(region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load config with temporary credentials: %w", err)
	}

	return cfg, nil
}

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

	fmt.Println("Instance launched successfully")

	config, err := CreateAWSConfigWithTempCredentials(&creds, region)
	if err != nil {
		log.Fatalf("Error creating AWS config with temporary credentials: %v", err)
	}

	err = DeleteTable(config, destinationTable)
	if err != nil {
		log.Fatalf("Error deleting table: %v", err)
	}

	err = WaitForDeleteTable(config, destinationTable)
	if err != nil {
		log.Fatalf("Error waiting for table deletion: %v", err)
	}

	fmt.Println("Table deleted successfully")

	err = DeleteInstance(creds, "us-east-1", "my-instance")
	if err != nil {
		log.Fatalf("Error deleting instance: %v", err)
	}

	svc := ec2.NewFromConfig(config)
	err = WaitForTerminateInstance(context.Background(), svc, "my-instance")
	if err != nil {
		log.Fatalf("Error waiting for instance termination: %v", err)
	}

	fmt.Println("Instance deleted successfully")

}
