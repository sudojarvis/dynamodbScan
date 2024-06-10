package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func DeleteInstance(creds types.Credentials, region, instanceName string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(
		awscredentials.NewStaticCredentialsProvider(aws.ToString(creds.AccessKeyId), aws.ToString(creds.SecretAccessKey), aws.ToString(creds.SessionToken)),
	), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	svc := ec2.NewFromConfig(cfg)

	instanceID, err := GetInstanceIDByName(svc, instanceName)
	if err != nil {
		return err
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err = svc.TerminateInstances(context.TODO(), input)

	if err != nil {
		return err
	}

	return nil
}

func WaitForTerminateInstance(ctx context.Context, svc *ec2.Client, instanceName string) error {
	maxAttempts := 10
	delay := time.Second * 30

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check instance status
		instanceID, err := GetInstanceIDByName(svc, instanceName)
		if err != nil {
			return err
		}
		describeOutput, err := svc.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		})

		if err != nil {
			return err
		}

		// Check if instance is terminated
		for _, reservation := range describeOutput.Reservations {
			for _, instance := range reservation.Instances {
				if *instance.InstanceId == instanceID && instance.State.Name == "terminated" {
					return nil // Instance terminated
				}
			}
		}

		// If instance is not terminated, wait and retry
		time.Sleep(delay)
	}

	return fmt.Errorf("instance did not terminate within the allotted time")
}

func GetInstanceIDByName(svc *ec2.Client, instanceName string) (string, error) {
	input := &ec2.DescribeInstancesInput{}

	result, err := svc.DescribeInstances(context.TODO(), input)
	if err != nil {
		return "", err
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" && *tag.Value == instanceName {
					return *instance.InstanceId, nil
				}
			}
		}
	}

	return "", fmt.Errorf("instance not found: %s", instanceName)
}
