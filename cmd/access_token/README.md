# Get access token Firebase tool (for testing)

Used to fetch an access token from Firebase for API authentication.

### Prerequisites

Firebase Web API Key from the Firebase Console → Project Settings → Web API Key.

### Usage

From the terminal, run: 

```bash
go run cmd/access_token/main.go -api-key=<firebase_web_api_key> -email=<email> -password=<password>
```

Get verbose output (includes user ID, expiration, refresh token):
```bash
go run cmd/access_token/main.go -api-key=<firebase_web_api_key> -email=<email> -password=<password> -verbose
```

Use the token in a curl request:
```bash
TOKEN=$(go run cmd/access_token/main.go -api-key=AIza... -email=test@example.com -password=pass123)
curl -X POST http://localhost:3000/v2/course \
  -H "Content-Type: application/json" \
  -d "{\"access_token\": \"$TOKEN\", \"courseId\": \"course-123\"}"
```