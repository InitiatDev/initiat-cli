# InitFlow Workspace API Specification

This document provides comprehensive documentation for the InitFlow Workspace API, including cryptographic key management and access control.

## Overview

The Workspace API manages workspace access and cryptographic key initialization for authenticated devices. Workspaces are containers for secrets and are associated with organizations for access control.

**Base URL**: `/api/v1`  
**Authentication**: Required (Device-based Ed25519 signatures)  
**Content-Type**: `application/json`

## Authentication

All workspace endpoints require device authentication. See [API_AUTHENTICATION_SPEC.md](./API_AUTHENTICATION_SPEC.md) for complete authentication details.

**Required Headers:**
```http
Authorization: Device {device_id}
X-Signature: {signature}
X-Timestamp: {timestamp}
```

## Endpoints

### GET /workspaces

Lists all workspaces accessible by the authenticated device's user.

#### Request

```http
GET /api/v1/workspaces
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
```

#### Response (200 OK)

```json
{
  "workspaces": [
    {
      "id": 42,
      "name": "Production Environment",
      "slug": "production",
      "description": "Production secrets and configuration",
      "key_initialized": true,
      "key_version": 1,
      "organization": {
        "id": 1,
        "name": "Acme Corp",
        "slug": "acme-corp"
      }
    },
    {
      "id": 43,
      "name": "Development Environment", 
      "slug": "development",
      "description": "Development secrets and configuration",
      "key_initialized": false,
      "key_version": null,
      "organization": {
        "id": 1,
        "name": "Acme Corp",
        "slug": "acme-corp"
      }
    }
  ]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Unique workspace identifier |
| `name` | string | Human-readable workspace name |
| `slug` | string | URL-friendly workspace identifier |
| `description` | string | Workspace description (nullable) |
| `key_initialized` | boolean | Whether workspace key has been initialized |
| `key_version` | integer | Current workspace key version (null if not initialized) |
| `organization.id` | integer | Parent organization ID |
| `organization.name` | string | Organization name |
| `organization.slug` | string | Organization slug |

#### Access Control

A device has access to a workspace if:
1. The device's user is a member of the workspace's organization
2. The device has a workspace key for this workspace (for key-initialized workspaces)

#### Error Responses

**401 Unauthorized** - Invalid authentication
```json
{
  "error": {
    "message": "Invalid signature"
  }
}
```

### POST /workspaces/:workspace_id/initialize

Initializes the workspace key for a workspace. This is a critical security operation that establishes the cryptographic foundation for secret storage.

#### Request

```http
POST /api/v1/workspaces/42/initialize
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
Content-Type: application/json

{
  "wrapped_workspace_key": "hSDwCYkwp1R0i33ctD73Wg2_Og0mOBr066SpjqqbTmo..."
}
```

#### Request Fields

| Field | Type | Description | Encoding |
|-------|------|-------------|----------|
| `wrapped_workspace_key` | string | Workspace key encrypted with device's X25519 public key | URL-Safe Base64 |

#### Cryptographic Flow

1. **Client generates** a random 32-byte workspace key (WSK)
2. **Client encrypts** WSK using device's X25519 public key
3. **Client sends** wrapped WSK to server
4. **Server stores** wrapped WSK associated with device and workspace
5. **Server marks** workspace as key-initialized

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Workspace key initialized successfully",
  "workspace": {
    "id": 42,
    "name": "Production Environment",
    "slug": "production", 
    "key_initialized": true,
    "key_version": 1
  }
}
```

#### Access Control

Only workspace owners can initialize keys. The server verifies:
1. Device authentication is valid
2. Device's user owns the workspace
3. Workspace is not already initialized

#### Error Responses

**403 Forbidden** - Not workspace owner
```json
{
  "error": {
    "message": "Only workspace owners can initialize keys"
  }
}
```

**409 Conflict** - Already initialized
```json
{
  "error": {
    "message": "Workspace key already initialized"
  }
}
```

**404 Not Found** - Workspace doesn't exist
```json
{
  "error": {
    "message": "Workspace not found"
  }
}
```

**422 Unprocessable Entity** - Validation errors
```json
{
  "success": false,
  "error": "Validation failed",
  "errors": {
    "wrapped_workspace_key": ["can't be blank"]
  }
}
```

## Cryptographic Details

### Workspace Key (WSK)

- **Size**: 32 bytes (256 bits)
- **Generation**: Cryptographically secure random
- **Purpose**: Encrypts/decrypts all secrets in the workspace
- **Storage**: Never stored in plaintext on server
- **Distribution**: Wrapped with each device's X25519 public key

### Key Wrapping

The workspace key is encrypted using X25519 key exchange:

**Pseudocode:**
```
ephemeral_private_key, ephemeral_public_key = generate_x25519_keypair()
shared_secret = x25519_key_exchange(device_public_key, ephemeral_private_key)
nonce = generate_random_24_bytes()
encrypted_workspace_key = xsalsa20_poly1305_encrypt(workspace_key, nonce, shared_secret)
wrapped_key = ephemeral_public_key + nonce + encrypted_workspace_key
```

**Process:**
1. Generate ephemeral X25519 keypair for this wrapping operation
2. Perform X25519 key exchange with device's public key
3. Generate random 24-byte nonce for encryption
4. Encrypt workspace key using XSalsa20-Poly1305 with shared secret
5. Concatenate ephemeral public key, nonce, and encrypted key

### Key Unwrapping

Devices unwrap the workspace key using their X25519 private key:

**Pseudocode:**
```
ephemeral_public_key = wrapped_key[0:32]
nonce = wrapped_key[32:56]
encrypted_workspace_key = wrapped_key[56:]
shared_secret = x25519_key_exchange(ephemeral_public_key, device_private_key)
workspace_key = xsalsa20_poly1305_decrypt(encrypted_workspace_key, nonce, shared_secret)
```

**Process:**
1. Extract ephemeral public key (first 32 bytes)
2. Extract nonce (next 24 bytes)
3. Extract encrypted workspace key (remaining bytes)
4. Perform X25519 key exchange with ephemeral public key
5. Decrypt workspace key using XSalsa20-Poly1305 with shared secret

## Security Considerations

### Zero-Knowledge Architecture

- **Server never sees**: Plaintext workspace keys or secret values
- **Server stores**: Only wrapped keys and encrypted secrets
- **Client responsibility**: All encryption/decryption operations

### Access Control Model

1. **Organization membership**: Controls workspace visibility
2. **Workspace keys**: Controls secret access within workspace
3. **Device authentication**: Ensures request authenticity

### Key Rotation

- **Current**: Single workspace key per workspace
- **Future**: Support for key versioning and rotation
- **Migration**: Gradual rollout with backward compatibility

### Threat Model

**Protected against:**
- Server compromise (keys remain encrypted)
- Network interception (all data encrypted)
- Unauthorized access (device authentication required)

**Not protected against:**
- Client device compromise (keys accessible to malware)
- Malicious workspace owners (can access all workspace secrets)
- Side-channel attacks on client devices

## Implementation Guidelines

### Client Implementation

**Workspace Initialization Flow:**
```
1. workspace_key = generate_random_32_bytes()
2. device_public_key = get_device_x25519_public_key()
3. wrapped_key = wrap_workspace_key(workspace_key, device_public_key)
4. encoded_wrapped_key = url_safe_base64_encode(wrapped_key)
5. send_authenticated_request("POST", "/workspaces/{id}/initialize", {
     "wrapped_workspace_key": encoded_wrapped_key
   })
```

**Implementation Steps:**
1. Generate cryptographically secure 32-byte workspace key
2. Retrieve device's X25519 public key from local storage
3. Wrap workspace key using key exchange process
4. Encode wrapped key as URL-safe Base64 for transport
5. Send authenticated POST request to initialize endpoint

### Error Handling

**Error Response Mapping:**
```
HTTP 403 → "Insufficient permissions for workspace"
HTTP 404 → "Workspace not found"
HTTP 409 → "Workspace key already initialized"
HTTP 422 → "Validation failed" (check response body for details)
HTTP 401 → "Authentication failed"
```

**Error Handling Strategy:**
1. Check HTTP status code first
2. Parse JSON error response for detailed message
3. Handle specific error cases with appropriate user feedback
4. Log technical details for debugging
5. Provide actionable error messages to users

## Testing

### Test Scenarios

1. **List workspaces** - Verify access control
2. **Initialize key** - Test cryptographic flow
3. **Access control** - Verify permission checks
4. **Error handling** - Test all error conditions
5. **Key wrapping** - Verify cryptographic correctness

### Test Vectors

```json
{
  "workspace_key": "dGVzdF93b3Jrc3BhY2Vfa2V5XzMyX2J5dGVzX2xvbmc",
  "device_public_key": "hSDwCYkwp1R0i33ctD73Wg2_Og0mOBr066SpjqqbTmo",
  "expected_wrapped_key_length": 88
}
```

## Migration and Versioning

### API Versioning
- **Current**: v1
- **Backward compatibility**: Maintained within major versions
- **Deprecation**: 6-month notice for breaking changes

### Data Migration
- **Key initialization**: One-time operation per workspace
- **Schema changes**: Handled via database migrations
- **Client updates**: Gradual rollout with feature flags
