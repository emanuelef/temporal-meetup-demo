#!/bin/sh
echo "Init localstack s3"
awslocal s3 mb s3://scripts-local
awslocal s3 cp /files s3://scripts-local --recursive
