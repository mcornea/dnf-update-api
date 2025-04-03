#!/bin/bash
set -e

# Build the container image
podman build -t dnf-update-api .

# Generate a random token if not provided
if [ -z "$API_TOKEN" ]; then
    API_TOKEN=$(openssl rand -hex 16)
    echo "Generated API token: $API_TOKEN"
fi

# Set KERNEL_UPDATE_REBOOT to false by default if not provided
if [ -z "$KERNEL_UPDATE_REBOOT" ]; then
    KERNEL_UPDATE_REBOOT="false"
    echo "KERNEL_UPDATE_REBOOT not set, defaulting to: $KERNEL_UPDATE_REBOOT"
fi

# Run the container
podman run -d --name dnf-update-api \
    -p 8080:8080 \
    -e API_TOKEN="$API_TOKEN" \
    -e KERNEL_UPDATE_REBOOT="$KERNEL_UPDATE_REBOOT" \
    --privileged \
    --health-cmd "curl -s http://localhost:8080/health | grep OK" \
    --health-on-failure=kill \
    --health-retries=3 
    -v /:/host \
    dnf-update-api

echo "DNF Update API is running on port 8080"
echo "Use this token for API requests: $API_TOKEN"echo "Use this token for API requests: $API_TOKEN"