# Initiat API Authentication Specification

This document provides comprehensive documentation for the Initiat API authentication system, including cryptographic details and implementation requirements.

## Overview

Initiat uses a two-phase authentication system:
1. **User Authentication**: Email/password login to obtain device registration tokens
2. **Device Authentication**: Cryptographic signature-based authentication for API requests

All cryptographic data uses **URL-Safe Base64 encoding without padding** for consistency and compatibility.

## Phase 1: User Authentication

### POST /api/v1/auth/login

Authenticates a user with email and password to obtain a device registration token.

#### Request

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "user_password"
}
```

#### Response (Success - 200 OK)

```json
{
  "token": "xyz789abc123def456...",
  "user": {
    "id": 42,
    "email": "user@example.com",
    "name": "John",
    "surname": "Doe"
  }
}
```

#### Response (Error - 401 Unauthorized)

```json
{
  "message": "Invalid email or password"
}
```

#### Response (Error - 400 Bad Request)

```json
{
  "message": "Email and password are required"
}
```

### Token Details

- **Format**: URL-Safe Base64 without padding
- **Size**: 32 bytes (43 characters when encoded)
- **Expiration**: 1 hour from generation
- **Usage**: Single-use for device registration

## Phase 2: Device Registration

### POST /api/v1/devices

Registers a new device using the authentication token from Phase 1.

#### Request

```http
POST /api/v1/devices
Content-Type: application/json

{
  "token": "xyz789abc123def456...",
  "name": "My Laptop",
  "public_key_ed25519": "MEUCIQDx...",
  "public_key_x25519": "MCQCIG7..."
}
```

#### Request Fields

| Field | Type | Description | Encoding |
|-------|------|-------------|----------|
| `token` | string | Device registration token from login | URL-Safe Base64 |
| `name` | string | Human-readable device name | Plain text |
| `public_key_ed25519` | string | Ed25519 public key for signing | URL-Safe Base64 |
| `public_key_x25519` | string | X25519 public key for encryption | URL-Safe Base64 |

#### Response (Success - 201 Created)

```json
{
  "success": true,
  "device": {
    "id": "abc123def456ghi789jkl",
    "name": "My Laptop",
    "created_at": "2023-09-20T12:34:56Z"
  }
}
```

#### Response (Error - 401 Unauthorized)

```json
{
  "error": {
    "message": "Invalid or expired registration token"
  }
}
```

#### Response (Error - 400 Bad Request)

```json
{
  "error": {
    "message": "Invalid ed25519 public key format"
  }
}
```

#### Response (Error - 422 Unprocessable Entity)

```json
{
  "success": false,
  "error": "Validation failed",
  "errors": {
    "name": ["can't be blank"]
  }
}
```

### Device ID Generation

- **Format**: 16 random bytes encoded as URL-Safe Base64
- **Size**: 22 characters (no padding)
- **Uniqueness**: Cryptographically random, collision-resistant
- **Usage**: Device identifier in API requests

## Phase 3: API Request Authentication

All authenticated API requests use Ed25519 signature-based authentication.

### Authentication Headers

Every authenticated request must include these headers:

```http
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
```

### Header Details

| Header | Description | Format |
|--------|-------------|--------|
| `Authorization` | Device ID with "Device" prefix | `Device {device_id}` |
| `X-Signature` | Ed25519 signature of request | URL-Safe Base64 |
| `X-Timestamp` | Unix timestamp in seconds | Integer string |

### Signature Generation

#### Step 1: Build the Message to Sign

The signature message format is **body-agnostic** for reliability:

```
{METHOD}\n{PATH}{QUERY_STRING}\n{TIMESTAMP}
```

**Components:**
- `METHOD`: HTTP method (GET, POST, DELETE, etc.)
- `PATH`: Request path (e.g., `/api/v1/workspaces/42/secrets`)
- `QUERY_STRING`: Query parameters with leading `?` (empty string if none)
- `TIMESTAMP`: Unix timestamp in seconds

**Examples:**

```
GET /api/v1/workspaces:
GET\n/api/v1/workspaces\n1694612345

POST /api/v1/workspaces/42/secrets:
POST\n/api/v1/workspaces/42/secrets\n1694612345

GET /api/v1/workspaces?limit=10:
GET\n/api/v1/workspaces?limit=10\n1694612345
```

#### Step 2: Sign the Message

**Pseudocode:**
```
timestamp = current_unix_timestamp()
full_path = path + query_string_with_leading_question_mark
message = method + "\n" + full_path + "\n" + timestamp
signature_bytes = ed25519_sign(private_key, message_bytes)
encoded_signature = url_safe_base64_encode(signature_bytes)
```

**Process:**
1. Generate current Unix timestamp in seconds
2. Combine path with query string (include leading `?` if query exists)
3. Build message string with newline separators
4. Sign message bytes using Ed25519 private key
5. Encode signature as URL-safe Base64 without padding

### Signature Verification

The server verifies signatures using this process:

1. **Extract headers** and validate format
2. **Reconstruct message** using the same format
3. **Verify timestamp** (must be within 5 minutes)
4. **Verify signature** using the device's Ed25519 public key
5. **Check device access** for the requested resource

### Security Considerations

#### Timestamp Validation
- **Window**: 5 minutes (300 seconds)
- **Purpose**: Prevent replay attacks
- **Clock skew**: Servers should use NTP for accurate time

#### Signature Security
- **Algorithm**: Ed25519 (RFC 8032)
- **Key size**: 32 bytes (256 bits)
- **Signature size**: 64 bytes (512 bits)
- **Encoding**: URL-Safe Base64 without padding

#### Body-Agnostic Design
- **Rationale**: Eliminates JSON formatting inconsistencies
- **Benefits**: More reliable, simpler implementation
- **Trade-off**: No protection against body tampering (acceptable for encrypted secrets)

## Error Responses

### Authentication Errors

#### 401 Unauthorized - Invalid Device
```json
{
  "error": {
    "message": "Invalid device ID"
  }
}
```

#### 401 Unauthorized - Invalid Signature
```json
{
  "error": {
    "message": "Invalid signature"
  }
}
```

#### 401 Unauthorized - Timestamp Issues
```json
{
  "error": {
    "message": "Request timestamp too old"
  }
}
```

#### 403 Forbidden - Access Denied
```json
{
  "error": {
    "message": "Device does not have access to this workspace"
  }
}
```

## Implementation Checklist

### Client Implementation
- [ ] Generate Ed25519 keypair for device
- [ ] Store private key securely (keychain/keyring)
- [ ] Implement signature generation with body-agnostic message
- [ ] Use URL-Safe Base64 encoding for all cryptographic data
- [ ] Handle timestamp synchronization
- [ ] Implement proper error handling for auth failures

### Server Implementation
- [ ] Validate device ID format and existence
- [ ] Reconstruct signature message correctly
- [ ] Verify Ed25519 signatures
- [ ] Check timestamp window (5 minutes)
- [ ] Implement device access control
- [ ] Return consistent error responses

## Testing

### Test Vectors

#### Device Registration
```json
{
  "token": "dGVzdF90b2tlbl8zMl9ieXRlc19sb25nX2Zvcl90ZXN0aW5n",
  "name": "Test Device",
  "public_key_ed25519": "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo",
  "public_key_x25519": "hSDwCYkwp1R0i33ctD73Wg2_Og0mOBr066SpjqqbTmo"
}
```

#### Signature Generation
```
Message: GET\n/api/v1/workspaces\n1694612345
Private Key: (32 bytes)
Expected Signature: (64 bytes, URL-Safe Base64 encoded)
```

### Integration Tests
- User login flow
- Device registration flow  
- Authenticated API requests
- Error handling scenarios
- Timestamp edge cases
- Invalid signature detection
