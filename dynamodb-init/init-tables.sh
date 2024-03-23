#!/bin/bash

# Wait for DynamoDB to be ready
until aws --endpoint-url=http://localhost:8000 dynamodb list-tables; do
  echo "DynamoDB is not ready yet. Waiting..."
  sleep 2
done

# Create DynamoDB table
aws --endpoint-url=http://localhost:8000 dynamodb create-table --cli-input-json file://table-definition.json

# Wait for table to be created
echo "Waiting for table creation..."
sleep 5

# Populate DynamoDB table with item
aws --endpoint-url=http://localhost:8000 dynamodb put-item --table-name ExampleTable --item file://item.json

echo "DynamoDB tables created and populated successfully"
