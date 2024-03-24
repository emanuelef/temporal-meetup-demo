#!/bin/bash

echo "########### Creating profile ###########"
aws configure set aws_access_key_id default_access_key --profile=$AWS_DEFAULT_PROFILE
aws configure set aws_secret_access_key default_secret_key --profile=$AWS_DEFAULT_PROFILE
aws configure set region us-east-1 --profile=$AWS_DEFAULT_PROFILE

aws configure list

# Wait for DynamoDB to be ready
until aws $AWS_ENDPOINT dynamodb list-tables; do
  echo "DynamoDB is not ready yet. Waiting..."
  sleep 2
done

# Create DynamoDB table
aws $AWS_ENDPOINT dynamodb create-table --cli-input-json file:///init-scripts/table-definition.json

# Wait for table to be created
echo "Waiting for table creation..."
sleep 5

# Populate DynamoDB table with item
aws $AWS_ENDPOINT dynamodb put-item --table-name Services --item file:///init-scripts/items.json

echo "DynamoDB tables created and populated successfully"
