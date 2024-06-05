package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func createTable(ctx context.Context, client *dynamodb.Client, sourceTable, destinationTable string) error {
	desc, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &sourceTable,
	})
	if err != nil {
		return fmt.Errorf("failed to describe source table: %w", err)
	}

	createInput := &dynamodb.CreateTableInput{
		TableName:            &destinationTable,
		KeySchema:            desc.Table.KeySchema,
		AttributeDefinitions: desc.Table.AttributeDefinitions,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  desc.Table.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: desc.Table.ProvisionedThroughput.WriteCapacityUnits,
		},
	}

	_, err = client.CreateTable(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create target table: %w", err)
	}

	// Wait for the table to become active
	for {
		desc, err = client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: &destinationTable,
		})
		if err != nil {
			return fmt.Errorf("failed to describe destination table: %w", err)
		}
		if desc.Table.TableStatus == types.TableStatusActive {
			break
		}
		log.Printf("Waiting for table %s to become active...", destinationTable)
		time.Sleep(5 * time.Second) // Sleep for 5 seconds before checking again
	}

	return nil
}

func scanSegment(ctx context.Context, client *dynamodb.Client, segment, totalSegments int32, sourceTable string) ([]map[string]types.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName:     &sourceTable,
		Segment:       &segment,
		TotalSegments: &totalSegments,
	}
	result, err := client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan segment %d: %w", segment, err)
	}
	return result.Items, nil
}

func writeItems(ctx context.Context, client *dynamodb.Client, items []map[string]types.AttributeValue, destinationTable string) error {
	for _, item := range items {
		input := &dynamodb.PutItemInput{
			TableName: &destinationTable,
			Item:      item,
		}
		_, err := client.PutItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to write item to table %s: %w", destinationTable, err)
		}
	}
	return nil
}

func parallelCopy(ctx context.Context, creds ststypes.Credentials, sourceTable, destinationTable, region string, totalSegments int) {
	client := dynamodb.NewFromConfig(aws.Config{
		Credentials: aws.NewCredentialsCache(awscredentials.NewStaticCredentialsProvider(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken)),
		Region:      region,
	})

	// Create the destination table
	err := createTable(ctx, client, sourceTable, destinationTable)
	if err != nil {
		log.Fatalf("unable to create destination table: %v", err)
	}
	log.Printf("Table %s created successfully.", destinationTable)

	var wg sync.WaitGroup
	errCh := make(chan error, totalSegments)

	for segment := 0; segment < totalSegments; segment++ {
		wg.Add(1)
		go func(segment int) {
			defer wg.Done()
			items, err := scanSegment(ctx, client, int32(segment), int32(totalSegments), sourceTable)
			if err != nil {
				errCh <- err
				return
			}
			if err := writeItems(ctx, client, items, destinationTable); err != nil {
				errCh <- err
				return
			}
		}(segment)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		log.Println("error during parallel copy:", err)
	}
}
