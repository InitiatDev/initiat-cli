# Initiat Workspace API Specification

This document provides comprehensive documentation for the Initiat Workspace API, including cryptographic key management and access control.

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
  "success": true,
  "data": {
    "workspaces": [
      {
        "id": 42,
        "name": "Production Environment",
        "slug": "production",
        "composite_slug": "acme-corp/production",
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
        "composite_slug": "acme-corp/development",
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
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Unique workspace identifier |
| `name` | string | Human-readable workspace name |
| `slug` | string | URL-friendly workspace identifier (unique within organization) |
| `composite_slug` | string | Globally unique identifier in format `org-slug/workspace-slug` |
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
  "success": false,
  "message": "Invalid signature",
  "errors": ["Invalid signature"]
}
```

### POST /workspaces/:org_slug/:workspace_slug/initialize

Initializes the workspace key for a workspace. This is a critical security operation that establishes the cryptographic foundation for secret storage.

#### Request

```http
POST /api/v1/workspaces/acme-corp/production/initialize
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
  "data": {
    "workspace": {
      "id": 42,
      "name": "Production Environment",
      "slug": "production", 
      "key_initialized": true,
      "key_version": 1
    }
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
  "success": false,
  "message": "Only workspace owners can initialize keys",
  "errors": ["Only workspace owners can initialize keys"]
}
```

**409 Conflict** - Already initialized
```json
{
  "success": false,
  "message": "Workspace key already initialized",
  "errors": ["Workspace key already initialized"]
}
```

**404 Not Found** - Workspace doesn't exist
```json
{
  "success": false,
  "message": "Workspace not found",
  "errors": ["Workspace not found"]
}
```

**404 Not Found** - Organization doesn't exist (composite slug format)
```json
{
  "success": false,
  "message": "Organization 'invalid-org' not found",
  "errors": ["Organization 'invalid-org' not found"]
}
```

**404 Not Found** - Workspace doesn't exist in organization (composite slug format)
```json
{
  "success": false,
  "message": "Workspace 'invalid-workspace' not found in organization 'acme-corp'",
  "errors": ["Workspace 'invalid-workspace' not found in organization 'acme-corp'"]
}
```

**422 Unprocessable Entity** - Validation errors
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "wrapped_workspace_key": ["can't be blank"]
  }
}
```

### GET /workspaces/:org_slug/:workspace_slug/workspace_key

Retrieves the wrapped workspace key for the authenticated device. This enables zero-persistence architecture where devices fetch and unwrap keys on-demand rather than storing them locally.

#### Request

```http
GET /api/v1/workspaces/acme-corp/production/workspace_key
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
```

#### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "wrapped_workspace_key": "hSDwCYkwp1R0i33ctD73Wg2_Og0mOBr066SpjqqbTmo...",
    "key_version": 1
  }
}
```

#### Response Fields

| Field | Type | Description | Encoding |
|-------|------|-------------|----------|
| `wrapped_workspace_key` | string | Workspace key encrypted for this device | URL-Safe Base64 |
| `key_version` | integer | Version of the workspace key |

#### Access Control

- Device must be authenticated
- Device must be approved for this workspace
- Returns wrapped key specific to the requesting device

#### Error Responses

**401 Unauthorized** - Invalid authentication
```json
{
  "success": false,
  "message": "Invalid signature",
  "errors": ["Invalid signature"]
}
```

**403 Forbidden** - Device not approved for workspace
```json
{
  "success": false,
  "message": "Device not approved for this workspace",
  "errors": ["Device not approved for this workspace"]
}
```

**404 Not Found** - Workspace not found or not initialized
```json
{
  "success": false,
  "message": "Workspace key not initialized",
  "errors": ["Workspace key not initialized"]
}
```

## Cryptographic Details

### Workspace Key (WSK)

- **Size**: 32 bytes (256 bits)
- **Generation**: Cryptographically secure random
- **Purpose**: Encrypts/decrypts all secrets in the workspace
- **Server Storage**: Never stored in plaintext on server; only wrapped versions stored
- **Client Storage**: Zero-persistence architecture - fetched and unwrapped on-demand
- **Distribution**: Wrapped with each device's X25519 public key

### Key Wrapping

The workspace key is encrypted using X25519 ECDH key exchange with ChaCha20-Poly1305:

**Pseudocode:**
```
ephemeral_private_key, ephemeral_public_key = generate_x25519_keypair()
shared_secret = x25519_key_exchange(device_public_key, ephemeral_private_key)
encryption_key = hkdf(shared_secret, salt="initiat.wrap", info="workspace")
nonce = generate_random_12_bytes()
encrypted_workspace_key = chacha20_poly1305_encrypt(workspace_key, nonce, encryption_key)
wrapped_key = ephemeral_public_key + nonce + encrypted_workspace_key
```

**Process:**
1. Generate ephemeral X25519 keypair for this wrapping operation
2. Perform X25519 key exchange with device's public key
3. Derive encryption key using HKDF-SHA256
4. Generate random 12-byte nonce for ChaCha20-Poly1305
5. Encrypt workspace key using ChaCha20-Poly1305 with derived key
6. Concatenate ephemeral public key (32 bytes) + nonce (12 bytes) + ciphertext

### Key Unwrapping

Devices unwrap the workspace key using their X25519 private key:

**Pseudocode:**
```
ephemeral_public_key = wrapped_key[0:32]
nonce = wrapped_key[32:44]
encrypted_workspace_key = wrapped_key[44:]
shared_secret = x25519_key_exchange(ephemeral_public_key, device_private_key)
encryption_key = hkdf(shared_secret, salt="initiat.wrap", info="workspace")
workspace_key = chacha20_poly1305_decrypt(encrypted_workspace_key, nonce, encryption_key)
```

**Process:**
1. Extract ephemeral public key (first 32 bytes)
2. Extract nonce (next 12 bytes)
3. Extract encrypted workspace key (remaining bytes)
4. Perform X25519 key exchange with ephemeral public key
5. Derive encryption key using HKDF-SHA256 (same parameters as wrapping)
6. Decrypt workspace key using ChaCha20-Poly1305 with derived key

### Zero-Persistence Client Architecture

The CLI implements a zero-persistence model for workspace keys:

**Traditional Approach (NOT used):**
```
Init: Generate WSK → Wrap → Send to server → Store plaintext locally
Use:  Read plaintext WSK from local storage → Decrypt secret
```

**Zero-Persistence Approach (IMPLEMENTED):**
```
Init: Generate WSK → Wrap → Send to server → Discard WSK
Use:  Fetch wrapped WSK from server → Unwrap with device key → Decrypt secret → Discard WSK
```

**Benefits:**
- No persistent plaintext workspace keys on client devices
- Real-time access revocation (server returns 403 when device approval revoked)
- Audit trail of all workspace key access
- Defense in depth against local storage compromise
- No local state synchronization issues

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
