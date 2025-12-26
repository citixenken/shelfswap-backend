#!/bin/bash

# ShelfSwap Backend Deployment Script for Google Cloud Run
# This script builds and deploys the backend to Google Cloud Run

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${GCP_PROJECT_ID:-}"
SERVICE_NAME="shelfswap-backend"
REGION="${GCP_REGION:-us-central1}"
IMAGE_NAME="gcr.io/${PROJECT_ID}/${SERVICE_NAME}"

echo -e "${GREEN}🚀 ShelfSwap Backend Deployment${NC}"
echo "=================================="

# Check if PROJECT_ID is set
if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: GCP_PROJECT_ID environment variable is not set${NC}"
    echo "Please set it with: export GCP_PROJECT_ID=your-project-id"
    exit 1
fi

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}Error: gcloud CLI is not installed${NC}"
    echo "Please install it from: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Check if user is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" &> /dev/null; then
    echo -e "${YELLOW}Authenticating with Google Cloud...${NC}"
    gcloud auth login
fi

# Set the project
echo -e "${YELLOW}Setting project to ${PROJECT_ID}...${NC}"
gcloud config set project ${PROJECT_ID}

# Enable required APIs
echo -e "${YELLOW}Enabling required APIs...${NC}"
gcloud services enable cloudbuild.googleapis.com run.googleapis.com containerregistry.googleapis.com

# Build the Docker image
echo -e "${YELLOW}Building Docker image...${NC}"
docker build -t ${IMAGE_NAME}:latest .

# Configure Docker to use gcloud as a credential helper
echo -e "${YELLOW}Configuring Docker authentication...${NC}"
gcloud auth configure-docker gcr.io --quiet

# Push the image to Google Container Registry
echo -e "${YELLOW}Pushing image to Google Container Registry...${NC}"
docker push ${IMAGE_NAME}:latest

# Deploy to Cloud Run
echo -e "${YELLOW}Deploying to Cloud Run...${NC}"
gcloud run deploy ${SERVICE_NAME} \
    --image ${IMAGE_NAME}:latest \
    --platform managed \
    --region ${REGION} \
    --allow-unauthenticated \
    --port 8080 \
    --memory 512Mi \
    --cpu 1 \
    --min-instances 0 \
    --max-instances 10 \
    --timeout 300

# Get the service URL
SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} --platform managed --region ${REGION} --format 'value(status.url)')

echo ""
echo -e "${GREEN}✅ Deployment successful!${NC}"
echo "=================================="
echo -e "Service URL: ${GREEN}${SERVICE_URL}${NC}"
echo ""
echo -e "${YELLOW}⚠️  Important Next Steps:${NC}"
echo "1. Set environment variables in Cloud Run console:"
echo "   - DATABASE_URL (Supabase PostgreSQL connection string)"
echo "   - SUPABASE_URL (Supabase project URL)"
echo "   - SUPABASE_SERVICE_ROLE_KEY (Supabase service role key)"
echo "   - RESEND_API_KEY (Resend email API key)"
echo ""
echo "   Run this command to set secrets:"
echo "   gcloud run services update ${SERVICE_NAME} --region ${REGION} \\"
echo "     --set-env-vars DATABASE_URL=your_database_url \\"
echo "     --set-env-vars SUPABASE_URL=your_supabase_url \\"
echo "     --set-env-vars SUPABASE_SERVICE_ROLE_KEY=your_supabase_key \\"
echo "     --set-env-vars RESEND_API_KEY=your_resend_key"
echo ""
echo "2. Update frontend _redirects file with this URL:"
echo "   /api/*  ${SERVICE_URL}/:splat  200"
echo ""
echo "3. Test the health endpoint:"
echo "   curl ${SERVICE_URL}/health"
echo ""
