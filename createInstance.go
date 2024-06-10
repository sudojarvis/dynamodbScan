// package createinstance
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

// func AssumeRole(roleArn, roleSessionName string) (ststypes.Credentials, error) {
// 	cfg, err := config.LoadDefaultConfig(context.TODO())
// 	if err != nil {
// 		return ststypes.Credentials{}, fmt.Errorf("unable to load SDK config: %v", err)
// 	}

// 	stsClient := sts.NewFromConfig(cfg)
// 	input := &sts.AssumeRoleInput{
// 		RoleArn:         aws.String(roleArn),
// 		RoleSessionName: aws.String(roleSessionName),
// 	}

// 	// result, err := stsClient.AssumeRole(context.TODO(), input)
// 	result, err := stsClient.AssumeRole(context.TODO(), input)
// 	if err != nil {
// 		return ststypes.Credentials{}, fmt.Errorf("error assuming role: %v", err)
// 	}

// 	return *result.Credentials, nil
// }

// func assumeRole(accessKey, secretKey, roleArn, roleSessionName string) (ststypes.Credentials, error) {
// 	cfg, err := config.LoadDefaultConfig(context.TODO(),
// 		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
// 	if err != nil {
// 		return ststypes.Credentials{}, fmt.Errorf("unable to load SDK config: %v", err)
// 	}

// 	stsClient := sts.NewFromConfig(cfg)
// 	input := &sts.AssumeRoleInput{
// 		RoleArn:         aws.String(roleArn),
// 		RoleSessionName: aws.String(roleSessionName),
// 	}

// 	result, err := stsClient.AssumeRole(context.TODO(), input)
// 	if err != nil {
// 		return ststypes.Credentials{}, fmt.Errorf("error assuming role: %v", err)
// 	}

// 	return *result.Credentials, nil
// }

func GetCallerIdentity(creds types.Credentials) (*sts.GetCallerIdentityOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(
		aws.NewCredentialsCache(
			awscredentials.NewStaticCredentialsProvider(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config with temporary credentials: %v", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	input := &sts.GetCallerIdentityInput{}

	result, err := stsClient.GetCallerIdentity(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error getting caller identity: %v", err)
	}

	return result, nil
}

func LaunchEC2Instance(creds types.Credentials, region string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region), config.WithCredentialsProvider(
		aws.NewCredentialsCache(
			awscredentials.NewStaticCredentialsProvider(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken),
		),
	))
	if err != nil {
		return fmt.Errorf("unable to load SDK config with temporary credentials: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-00beae93a2d981137"),
		InstanceType: "t2.micro",
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	result, err := ec2Client.RunInstances(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error launching instance: %v", err)
	}

	instanceID := *result.Instances[0].InstanceId
	waiter := ec2.NewInstanceRunningWaiter(ec2Client)
	err = waiter.Wait(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("error waiting for instance to run: %v", err)
	}

	fmt.Println("Instance launched successfully:", instanceID)
	return nil
}
