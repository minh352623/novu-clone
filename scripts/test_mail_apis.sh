#!/bin/bash

# Mail Module API Test Script
# Base URL - Change this to your actual server URL
BASE_URL="http://localhost:8080/v1/api"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}    Mail Module API Test Script${NC}"
echo -e "${YELLOW}========================================${NC}"

# ============================================
# 1. Get All Mail Jobs (with pagination)
# ============================================
echo -e "\n${GREEN}1. Get All Mail Jobs${NC}"
echo "GET ${BASE_URL}/mails/jobs"
curl --location --request GET "${BASE_URL}/mails/jobs?limit=10" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 2. Get All Mail Jobs with status filter
# ============================================
echo -e "\n${GREEN}2. Get All Mail Jobs (filter by status=completed)${NC}"
echo "GET ${BASE_URL}/mails/jobs?status=completed"
curl --location --request GET "${BASE_URL}/mails/jobs?limit=10&status=completed" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 3. Get Overall Mail Statistics
# ============================================
echo -e "\n${GREEN}3. Get Overall Mail Statistics${NC}"
echo "GET ${BASE_URL}/mails/stats"
curl --location --request GET "${BASE_URL}/mails/stats" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 4. Get Single Mail Job Status
# ============================================
JOB_ID=1
echo -e "\n${GREEN}4. Get Mail Job Status (Job ID: ${JOB_ID})${NC}"
echo "GET ${BASE_URL}/mails/jobs/${JOB_ID}"
curl --location --request GET "${BASE_URL}/mails/jobs/${JOB_ID}" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 5. Get Mail Job Statistics
# ============================================
echo -e "\n${GREEN}5. Get Mail Job Statistics (Job ID: ${JOB_ID})${NC}"
echo "GET ${BASE_URL}/mails/jobs/${JOB_ID}/stats"
curl --location --request GET "${BASE_URL}/mails/jobs/${JOB_ID}/stats" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 6. Get Mail Job Send Logs (with pagination)
# ============================================
echo -e "\n${GREEN}6. Get Mail Job Send Logs (Job ID: ${JOB_ID})${NC}"
echo "GET ${BASE_URL}/mails/jobs/${JOB_ID}/logs"
curl --location --request GET "${BASE_URL}/mails/jobs/${JOB_ID}/logs?limit=20" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 7. Get Mail Job Send Logs (filter by failed)
# ============================================
echo -e "\n${GREEN}7. Get Mail Job Send Logs - Failed Only (Job ID: ${JOB_ID})${NC}"
echo "GET ${BASE_URL}/mails/jobs/${JOB_ID}/logs?status=failed"
curl --location --request GET "${BASE_URL}/mails/jobs/${JOB_ID}/logs?limit=20&status=failed" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 8. Get Mail Job Tracking Stats
# ============================================
echo -e "\n${GREEN}8. Get Mail Job Tracking Stats (Job ID: ${JOB_ID})${NC}"
echo "GET ${BASE_URL}/mails/jobs/${JOB_ID}/tracking"
curl --location --request GET "${BASE_URL}/mails/jobs/${JOB_ID}/tracking" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 9. Retry Failed Emails
# ============================================
echo -e "\n${GREEN}9. Retry Failed Emails (Job ID: ${JOB_ID})${NC}"
echo "POST ${BASE_URL}/mails/jobs/${JOB_ID}/retry"
curl --location --request POST "${BASE_URL}/mails/jobs/${JOB_ID}/retry" \
--header 'accept: application/json' \
--header 'content-type: application/json'

echo -e "\n"

# ============================================
# 10. Send Mail Batch (requires file upload)
# ============================================
echo -e "\n${GREEN}10. Send Mail Batch (Example - requires actual file)${NC}"
echo "POST ${BASE_URL}/mails/send-batch"
echo -e "${YELLOW}Note: This requires a multipart form with template_id and file${NC}"
echo "Example:"
echo 'curl --location --request POST "${BASE_URL}/mails/send-batch" \'
echo '  --form "template_id=1" \'
echo '  --form "file=@/path/to/recipients.xlsx" \'
echo '  --form "scheduled_at=2026-01-28 10:00:00"  # Optional: for scheduled sending'

echo -e "\n"

# ============================================
# 11. Track Email Open (Tracking Pixel)
# ============================================
TRACKING_TOKEN="your_tracking_token_here"
echo -e "\n${GREEN}11. Track Email Open (Tracking Pixel)${NC}"
echo "GET ${BASE_URL}/track/${TRACKING_TOKEN}"
echo -e "${YELLOW}Note: This returns a 1x1 transparent PNG image${NC}"
echo "curl --location --request GET '${BASE_URL}/track/${TRACKING_TOKEN}' --output pixel.png"

echo -e "\n"
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}    Test Script Completed${NC}"
echo -e "${YELLOW}========================================${NC}"
