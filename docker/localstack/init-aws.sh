#!/bin/sh
# Runs inside LocalStack once it is ready (mounted into /etc/localstack/init/ready.d).
# Creates the AWS resources the server and subscriber expect; see .env.example.
set -eu

REGION="${AWS_DEFAULT_REGION:-ap-northeast-1}"

awslocal s3 mb "s3://go-ddd-template-bucket" --region "$REGION"

TOPIC_ARN=$(awslocal sns create-topic --name user-events --region "$REGION" --query TopicArn --output text)
QUEUE_URL=$(awslocal sqs create-queue --queue-name user-events --region "$REGION" --query QueueUrl --output text)
QUEUE_ARN=$(awslocal sqs get-queue-attributes --queue-url "$QUEUE_URL" --attribute-names QueueArn \
  --region "$REGION" --query Attributes.QueueArn --output text)

# Raw delivery: the SQS body is the CloudEvent JSON itself, not an SNS envelope.
awslocal sns subscribe --topic-arn "$TOPIC_ARN" --protocol sqs --notification-endpoint "$QUEUE_ARN" \
  --attributes RawMessageDelivery=true --region "$REGION" > /dev/null

echo "init-aws: bucket go-ddd-template-bucket, topic $TOPIC_ARN -> queue $QUEUE_URL"
