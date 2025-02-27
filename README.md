
# API Usage Instructions

## Health Check
To check if the API is running, you can use the following curl command:


curl -X GET http://localhost:8080/health


## List Available Updates
To list available updates, use the following command:


curl -X GET http://localhost:8080/api/updates -H "Authorization: Bearer YOUR_API_TOKEN"


## Trigger System Upgrade
To trigger a system upgrade, use the following command:


curl -X POST http://localhost:8080/api/upgrade -H "Authorization: Bearer YOUR_API_TOKEN"

