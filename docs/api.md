# API

## Cryptogen

### Generate Assets

```txt
POST /crypto-assets
```

Request:

```json
{
  "PeerOrgs": [
    {
      "Name": "Org1", 
      "Domain": "org1.example.com", 
      "Template": 
      {
        "Count": 1
      }, 
      "Users":
      {
        "Count": 1
      }
    }
  ]
}
```

Response:

```txt
HTTP/1.1 202 Accepted
```

```json
{
  "JobID": "1538212567-25937"
}
```

### Get Job Status

Request:

```txt
GET /status/:jobID
```

Response:

```txt
HTTP/1.1 200 OK
```

```json
{
  "Status": "processing",
  "SecretPaths": []
}
```

```json
{
  "Status": "complete",
  "SecretPaths": [
    "/secret/cryptogen/peerOrganizations/org1.example.com/ca",
    "..."
  ]
}
```
