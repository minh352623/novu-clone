#!/bin/bash

# Quick Test Script for Scheduled Emails
# Usage: ./test_scheduled_emails.sh

echo "🧪 Testing Scheduled Emails Feature"
echo "===================================="

# Configuration
BASE_URL="http://localhost:8098"
TEMPLATE_ID=1
CSV_FILE="docs/sample_mail_recipients.csv"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}Step 1: Schedule email for 2 minutes from now${NC}"
echo "------------------------------------------------"

# Calculate scheduled time (2 minutes from now)
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    SCHEDULED=$(date -v+2M '+%Y-%m-%d %H:%M:%S')
else
    # Linux
    SCHEDULED=$(date -d '+2 minutes' '+%Y-%m-%d %H:%M:%S')
fi

echo "Scheduled time: $SCHEDULED"

# Send request
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/api/mails/send-batch" \
  -F "template_id=$TEMPLATE_ID" \
  -F "file=@$CSV_FILE" \
  -F "scheduled_at=$SCHEDULED")

echo "Response: $RESPONSE"

# Extract job ID
JOB_ID=$(echo $RESPONSE | grep -o '"job_id":[0-9]*' | grep -o '[0-9]*')

if [ -z "$JOB_ID" ]; then
    echo -e "${YELLOW}⚠️  Failed to get job ID${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Job created with ID: $JOB_ID${NC}"

echo ""
echo -e "${BLUE}Step 2: Check initial job status${NC}"
echo "-----------------------------------"

sleep 2
STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/api/mails/jobs/$JOB_ID")
echo "Status: $STATUS_RESPONSE"

echo ""
echo -e "${YELLOW}⏰ Waiting for scheduled time ($SCHEDULED)...${NC}"
echo "Scheduler will process the job when ready."
echo "You can check status with: curl -X GET '$BASE_URL/v1/api/mails/jobs/$JOB_ID'"

echo ""
echo -e "${BLUE}Step 3: Monitor job (checking every 30 seconds)${NC}"
echo "-------------------------------------------------"

# Monitor for up to 3 minutes
for i in {1..6}; do
    echo "Check $i/6..."
    STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/api/mails/jobs/$JOB_ID")
    
    # Extract status
    STATUS=$(echo $STATUS_RESPONSE | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    echo "Job status: $STATUS"
    
    if [ "$STATUS" = "completed" ]; then
        echo -e "${GREEN}✅ Job completed successfully!${NC}"
        echo "Final response: $STATUS_RESPONSE"
        exit 0
    elif [ "$STATUS" = "failed" ]; then
        echo -e "${YELLOW}⚠️  Job failed${NC}"
        echo "Response: $STATUS_RESPONSE"
        exit 1
    fi
    
    if [ $i -lt 6 ]; then
        sleep 30
    fi
done

echo ""
echo -e "${YELLOW}⏳ Job still processing. Check manually:${NC}"
echo "curl -X GET '$BASE_URL/v1/api/mails/jobs/$JOB_ID'"
