#!/bin/bash
echo "🚀 Deploying Gintugas to Koyeb..."

# Build locally first (optional)
echo "1. Building application..."
go build -o gintugas .

# Push to GitHub
echo "2. Pushing to GitHub..."
git add .
git commit -m "Deploy to Koyeb"
git push origin main

# Deploy to Koyeb
echo "3. Deploying to Koyeb..."
koyeb service redeploy gintugas/api

echo "✅ Deployment initiated!"
echo "📊 Check logs: koyeb service logs gintugas/api"
echo "🌐 Your app: https://gintugas-api-<your-org>.koyeb.app"