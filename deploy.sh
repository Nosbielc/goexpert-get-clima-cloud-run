#!/bin/bash

# Script para deploy no Google Cloud Run
# Usage: ./deploy.sh PROJECT_ID WEATHER_API_KEY

if [ $# -ne 2 ]; then
    echo "Usage: $0 PROJECT_ID WEATHER_API_KEY"
    echo "Example: $0 my-project-id abc123def456"
    exit 1
fi

PROJECT_ID=$1
WEATHER_API_KEY=$2
IMAGE_NAME="gcr.io/$PROJECT_ID/weather-api"

echo "🔨 Building and pushing Docker image..."
gcloud builds submit --tag $IMAGE_NAME --project $PROJECT_ID

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "🚀 Deploying to Cloud Run..."
gcloud run deploy weather-api \
    --image $IMAGE_NAME \
    --platform managed \
    --region us-central1 \
    --allow-unauthenticated \
    --set-env-vars WEATHER_API_KEY=$WEATHER_API_KEY \
    --set-env-vars PORT=8080 \
    --memory 512Mi \
    --cpu 1 \
    --max-instances 10 \
    --project $PROJECT_ID

if [ $? -eq 0 ]; then
    echo "✅ Deploy completed successfully!"
    echo "🌐 Getting service URL..."
    gcloud run services describe weather-api --platform managed --region us-central1 --format 'value(status.url)' --project $PROJECT_ID
else
    echo "❌ Deploy failed"
    exit 1
fi
