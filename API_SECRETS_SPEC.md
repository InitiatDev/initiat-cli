# InitFlow Secrets API Specification

This document provides comprehensive documentation for the InitFlow Secrets API, including cryptographic secret management and zero-knowledge architecture.

## Overview

The Secrets API manages encrypted secrets within workspaces. All secrets are encrypted client-side with the workspace key (WSK) before transmission. The server never sees plaintext secret values, maintaining a zero-knowledge architecture.

**Base URL**: `/api/v1`  
**Authentication**: Required (Device-based Ed25519 signatures)  
**Content-Type**: `application/json`  
**Encryption**: Client-side with workspace keys

## Authentication

All secret endpoints require device authentication. See [API_AUTHENTICATION_SPEC.md](./API_AUTHENTICATION_SPEC.md) for complete authentication details.

**Required Headers:**
```http
Authorization: Device {device_id}
X-Signature: {signature}
X-Timestamp: {timestamp}
```

## Endpoints

### GET /workspaces/:workspace_id/secrets

Lists all secrets in a workspace (metadata only, no encrypted values).

#### Request

```http
GET /api/v1/workspaces/42/secrets
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
```

#### Response (200 OK)

```json
{
  "secrets": [
    {
      "key": "DATABASE_URL",
      "version": 2,
      "workspace_id": 42,
      "created_at": "2023-09-20T12:34:56Z",
      "updated_at": "2023-09-20T14:22:10Z",
      "created_by_device": {
        "id": "abc123def456ghi789jkl",
        "name": "My Laptop"
      }
    },
    {
      "key": "API_SECRET_KEY",
      "version": 1,
      "workspace_id": 42,
      "created_at": "2023-09-20T10:15:30Z",
      "updated_at": "2023-09-20T10:15:30Z",
      "created_by_device": {
        "id": "def456ghi789jkl123abc",
        "name": "CI Server"
      }
    }
  ]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `key` | string | Secret identifier (unique within workspace) |
| `version` | integer | Current version number (increments on updates) |
| `workspace_id` | integer | Parent workspace ID |
| `created_at` | string | ISO 8601 timestamp of creation |
| `updated_at` | string | ISO 8601 timestamp of last update |
| `created_by_device.id` | string | Device ID that created this secret |
| `created_by_device.name` | string | Human-readable device name |

#### Access Control

Device must have access to the workspace:
1. Device's user is a member of the workspace's organization
2. Device has a workspace key for this workspace

### GET /workspaces/:workspace_id/secrets/:key

Retrieves a specific secret with encrypted value and nonce for client-side decryption.

#### Request

```http
GET /api/v1/workspaces/42/secrets/DATABASE_URL
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
```

#### Response (200 OK)

```json
{
  "secret": {
    "key": "DATABASE_URL",
    "version": 2,
    "workspace_id": 42,
    "encrypted_value": "dGVzdF9lbmNyeXB0ZWRfdmFsdWU",
    "nonce": "dGVzdF9ub25jZQ",
    "created_at": "2023-09-20T12:34:56Z",
    "updated_at": "2023-09-20T14:22:10Z",
    "created_by_device": {
      "id": "abc123def456ghi789jkl",
      "name": "My Laptop"
    }
  }
}
```

#### Response Fields

| Field | Type | Description | Encoding |
|-------|------|-------------|----------|
| `encrypted_value` | string | Encrypted secret value | URL-Safe Base64 |
| `nonce` | string | Encryption nonce | URL-Safe Base64 |
| *(other fields)* | - | Same as list endpoint | - |

#### Error Responses

**404 Not Found** - Secret doesn't exist
```json
{
  "error": {
    "message": "Secret not found"
  }
}
```

### POST /workspaces/:workspace_id/secrets

Creates a new secret or updates an existing one (creates new version).

#### Request

```http
POST /api/v1/workspaces/42/secrets
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
Content-Type: application/json

{
  "key": "NEW_API_KEY",
  "encrypted_value": "dGVzdF9lbmNyeXB0ZWRfdmFsdWU",
  "nonce": "dGVzdF9ub25jZQ"
}
```

#### Request Fields

| Field | Type | Description | Encoding | Constraints |
|-------|------|-------------|----------|-------------|
| `key` | string | Secret identifier | Plain text | 1-255 characters, unique within workspace |
| `encrypted_value` | string | Encrypted secret value | URL-Safe Base64 | Required |
| `nonce` | string | Encryption nonce | URL-Safe Base64 | Required |

#### Response (201 Created)

```json
{
  "secret": {
    "key": "NEW_API_KEY",
    "version": 1,
    "workspace_id": 42,
    "encrypted_value": "dGVzdF9lbmNyeXB0ZWRfdmFsdWU",
    "nonce": "dGVzdF9ub25jZQ",
    "created_at": "2023-09-20T15:30:45Z",
    "updated_at": "2023-09-20T15:30:45Z",
    "created_by_device": {
      "id": "abc123def456ghi789jkl",
      "name": "My Laptop"
    }
  }
}
```

#### Versioning Behavior

- **New secret**: Creates version 1
- **Existing secret**: Creates new version (increments version number)
- **Soft deletion**: Previous versions remain in database but marked as deleted

#### Error Responses

**400 Bad Request** - Invalid Base64 encoding
```json
{
  "error": {
    "message": "Invalid Base64 encoding for encrypted_value"
  }
}
```

**422 Unprocessable Entity** - Validation errors
```json
{
  "success": false,
  "error": "Validation failed",
  "errors": {
    "key": ["can't be blank"],
    "encrypted_value": ["can't be blank"]
  }
}
```

### DELETE /workspaces/:workspace_id/secrets/:key

Soft deletes a secret (marks as deleted, preserves for audit).

#### Request

```http
DELETE /api/v1/workspaces/42/secrets/OLD_API_KEY
Authorization: Device abc123def456ghi789jkl
X-Signature: MEUCIQDx...
X-Timestamp: 1694612345
```

#### Response (204 No Content)

Empty response body.

#### Error Responses

**404 Not Found** - Secret doesn't exist
```json
{
  "error": {
    "message": "Secret not found"
  }
}
```

## Cryptographic Details

### Client-Side Encryption

All secrets are encrypted client-side using the workspace key (WSK) before transmission.

#### Encryption Algorithm

- **Cipher**: ChaCha20-Poly1305 (AEAD)
- **Key**: 32-byte workspace key
- **Nonce**: 24 bytes (NaCl secretbox format)
- **Authentication**: Built-in with AEAD

#### Encryption Flow

**Pseudocode:**
```
nonce = generate_random_24_bytes()
encrypted_value = chacha20_poly1305_encrypt(plaintext, nonce, workspace_key)
```

**Process:**
1. Generate cryptographically secure 24-byte nonce
2. Encrypt plaintext using ChaCha20-Poly1305 AEAD
3. Return encrypted value and nonce for storage
4. Encode both as URL-safe Base64 for transport

#### Decryption Flow

**Pseudocode:**
```
plaintext = chacha20_poly1305_decrypt(encrypted_value, nonce, workspace_key)
if decryption_failed:
    return error("Invalid ciphertext or key")
```

**Process:**
1. Decode encrypted value and nonce from URL-safe Base64
2. Decrypt using ChaCha20-Poly1305 AEAD with workspace key
3. Verify authentication tag (automatic with AEAD)
4. Return plaintext or error if decryption fails

### Transport Encoding

All binary cryptographic data uses URL-Safe Base64 encoding without padding:

**Encoding for API Transport:**
```
encrypted_b64 = url_safe_base64_encode(encrypted_bytes)
nonce_b64 = url_safe_base64_encode(nonce_bytes)
```

**Decoding from API Response:**
```
encrypted_bytes = url_safe_base64_decode(encrypted_b64)
nonce_bytes = url_safe_base64_decode(nonce_b64)
```

**Encoding Properties:**
- Character set: `A-Z`, `a-z`, `0-9`, `-`, `_`
- No padding characters (`=`)
- Safe for URLs, headers, and JSON

### Nonce Requirements

- **Size**: 24 bytes (192 bits)
- **Generation**: Cryptographically secure random
- **Uniqueness**: Must be unique per encryption operation
- **Reuse**: Never reuse nonces with the same key

## Zero-Knowledge Architecture

### Server Responsibilities

The server acts as a "dumb pipe" for encrypted data:

1. **Stores**: Encrypted values and nonces as opaque binary data
2. **Validates**: Request authentication and access control
3. **Never sees**: Plaintext secret values or workspace keys
4. **Provides**: Metadata and encrypted data transport

### Client Responsibilities

Clients handle all cryptographic operations:

1. **Key management**: Derive and store workspace keys securely
2. **Encryption**: Encrypt secrets before sending to server
3. **Decryption**: Decrypt secrets after receiving from server
4. **Nonce generation**: Generate unique nonces for each encryption

### Security Properties

- **Confidentiality**: Server compromise doesn't expose secret values
- **Integrity**: AEAD encryption provides authenticity
- **Access control**: Device authentication controls access
- **Auditability**: All operations logged with device attribution

## Implementation Guidelines

### Client Implementation

**Secret Creation Flow:**
```
1. nonce = generate_random_24_bytes()
2. encrypted_value = chacha20_poly1305_encrypt(secret_plaintext, nonce, workspace_key)
3. encrypted_b64 = url_safe_base64_encode(encrypted_value)
4. nonce_b64 = url_safe_base64_encode(nonce)
5. send_authenticated_request("POST", "/workspaces/{id}/secrets", {
     "key": secret_key,
     "encrypted_value": encrypted_b64,
     "nonce": nonce_b64
   })
```

**Secret Retrieval Flow:**
```
1. response = send_authenticated_request("GET", "/workspaces/{id}/secrets/{key}")
2. encrypted_value = url_safe_base64_decode(response.encrypted_value)
3. nonce = url_safe_base64_decode(response.nonce)
4. plaintext = chacha20_poly1305_decrypt(encrypted_value, nonce, workspace_key)
5. return plaintext
```

**Implementation Requirements:**
- Store workspace key securely in local keychain/keyring
- Generate unique nonces for each encryption operation
- Handle decryption failures gracefully
- Clear sensitive data from memory after use
- Validate all cryptographic operations

### Error Handling

**Error Response Mapping:**
```
HTTP 400 → "Invalid request format or Base64 encoding"
HTTP 401 → "Authentication failed"
HTTP 403 → "Insufficient permissions for workspace"
HTTP 404 → "Secret not found"
HTTP 422 → "Validation failed" (check response body for field errors)
```

**Error Handling Strategy:**
1. Check HTTP status code for error category
2. Parse JSON error response for specific details
3. Handle cryptographic errors (decryption failures)
4. Provide user-friendly error messages
5. Log technical details for debugging

### Key Derivation

**Workspace Key Derivation Flow:**
```
1. wrapped_key_b64 = get_wrapped_key_from_server(workspace_id)
2. wrapped_key = url_safe_base64_decode(wrapped_key_b64)
3. workspace_key = unwrap_workspace_key(wrapped_key, device_private_key)
4. validate workspace_key.length == 32 bytes
5. store workspace_key securely for secret operations
```

**Process:**
1. Retrieve wrapped workspace key from server
2. Decode from URL-safe Base64 transport encoding
3. Unwrap using X25519 key exchange (see workspace spec)
4. Validate key is exactly 32 bytes
5. Store securely in memory for secret encryption/decryption

## Security Considerations

### Threat Model

**Protected against:**
- Server compromise (secrets remain encrypted)
- Network interception (end-to-end encryption)
- Unauthorized access (device authentication required)
- Data tampering (AEAD authentication)

**Not protected against:**
- Client device compromise (workspace keys accessible)
- Malicious workspace members (shared workspace key)
- Side-channel attacks on client devices
- Quantum attacks (post-quantum cryptography not implemented)

### Best Practices

#### Client-Side Security
- Store workspace keys in secure storage (keychain/keyring)
- Clear sensitive data from memory after use
- Validate all cryptographic operations
- Use secure random number generation

#### Operational Security
- Rotate workspace keys periodically
- Monitor access patterns for anomalies
- Implement proper logging and auditing
- Use secure communication channels

### Compliance Considerations

- **GDPR**: Right to deletion (soft delete preserves audit trail)
- **SOC 2**: Access controls and audit logging
- **HIPAA**: Encryption at rest and in transit
- **PCI DSS**: Strong cryptography and key management

## Testing

### Test Scenarios

1. **CRUD operations** - Create, read, update, delete secrets
2. **Encryption/decryption** - Verify cryptographic correctness
3. **Access control** - Test workspace permissions
4. **Error handling** - All error conditions
5. **Versioning** - Secret version management
6. **Encoding** - Base64 transport encoding

### Test Vectors

```json
{
  "plaintext": "database_password_123",
  "workspace_key": "dGVzdF93b3Jrc3BhY2Vfa2V5XzMyX2J5dGVzX2xvbmc",
  "nonce": "dGVzdF9ub25jZV8yNF9ieXRlc19sb25nX2Zvcl90ZXN0",
  "expected_encrypted": "...",
  "expected_b64_encrypted": "...",
  "expected_b64_nonce": "..."
}
```

### Integration Tests

**Secret Lifecycle Test Flow:**
```
1. CREATE: client.create_secret("TEST_KEY", "test_value")
   → Verify HTTP 201 response
   
2. LIST: secrets = client.list_secrets()
   → Verify "TEST_KEY" appears in list
   
3. GET: value = client.get_secret("TEST_KEY")
   → Verify decrypted value equals "test_value"
   
4. UPDATE: client.create_secret("TEST_KEY", "updated_value")
   → Verify new version created (version = 2)
   
5. DELETE: client.delete_secret("TEST_KEY")
   → Verify HTTP 204 response
   
6. VERIFY: client.get_secret("TEST_KEY")
   → Verify HTTP 404 error
```

**Test Categories:**
- **CRUD Operations**: Create, read, update, delete workflows
- **Encryption/Decryption**: Verify cryptographic correctness
- **Access Control**: Test workspace permission enforcement
- **Error Handling**: All error conditions and edge cases
- **Versioning**: Secret version management and history
- **Encoding**: Base64 transport encoding validation

## Migration and Versioning

### API Versioning
- **Current**: v1
- **Backward compatibility**: Maintained within major versions
- **Breaking changes**: New major version with migration path

### Cryptographic Agility
- **Current**: ChaCha20-Poly1305 with 24-byte nonces
- **Future**: Support for multiple cipher suites
- **Migration**: Gradual rollout with version indicators

### Data Migration
- **Secret versioning**: Automatic on updates
- **Key rotation**: Future feature for workspace keys
- **Format changes**: Handled via API versioning
