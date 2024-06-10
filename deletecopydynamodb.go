package main

import (
	"context"
	"fmt"
	"time"

	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// deleteTable deletes a DynamoDB table.
func DeleteTable(cfg aws.Config, tableName string) error {
	// Create a DynamoDB client
	svc := dynamodb.NewFromConfig(cfg)

	// Delete table
	_, err := svc.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete table %s: %w", tableName, err)
	}

	fmt.Printf("Delete request sent for table %s\n", tableName)
	return nil
}

// waitForDeleteTable waits until the DynamoDB table is deleted.
func WaitForDeleteTable(cfg aws.Config, tableName string) error {
	svc := dynamodb.NewFromConfig(cfg)
	maxAttempts := 10
	delay := time.Second * 30

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Describe the table to get its status
		output, err := svc.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})

		// Handle the case where the table is not found (it means the table is deleted)
		var notFoundErr *types.ResourceNotFoundException
		if ok := errors.As(err, &notFoundErr); ok {
			fmt.Printf("Table %s is deleted\n", tableName)
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to describe table %s: %w", tableName, err)
		}

		// Check if table status is still "DELETING"
		if output.Table.TableStatus == types.TableStatusDeleting {
			fmt.Printf("Table %s is still deleting, waiting...\n", tableName)
			time.Sleep(delay)
			continue
		}

		// Unexpected status if not deleting and not deleted
		return fmt.Errorf("unexpected status for table %s: %s", tableName, output.Table.TableStatus)
	}

	return fmt.Errorf("table %s was not deleted after %d attempts", tableName, maxAttempts)
}
