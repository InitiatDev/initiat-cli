# InitFlow Secret Management Specification

**Version:** 2.0  
**Date:** September 2025  
**Status:** Authoritative

---

## Overview

InitFlow implements a **zero-knowledge secret management system** where the server never has access to plaintext secret values, workspace keys, or user passwords. All encryption and decryption operations occur client-side using industry-standard cryptographic libraries.

## Architecture Principles

### 🔐 Zero-Knowledge Architecture
- **Server Role**: "Dumb pipe" that stores and returns encrypted data exactly as received
- **Client Role**: Handles all cryptographic operations (encryption, decryption, key generation)
- **Security Model**: Server compromise cannot expose plaintext secrets

### 🔑 Key Hierarchy
```
User Device Keys (Ed25519 + X25519)
    ↓ (encrypts)
Workspace Keys (32-byte symmetric keys)
    ↓ (encrypts)
Secret Values (arbitrary plaintext)
```

---

## Cryptographic Specifications

### Secret Encryption Algorithm
**Primary**: NaCl secretbox (XSalsa20Poly1305)
- **Library**: `golang.org/x/crypto/nacl/secretbox`
- **Key Size**: 32 bytes (256 bits)
- **Nonce Size**: 24 bytes (192 bits)
- **Authentication**: Built-in (Poly1305 MAC)
- **Security Level**: 256-bit equivalent

### Encoding Standards
**All cryptographic data**: URL-Safe Base64 without padding
- **Character Set**: `A-Z`, `a-z`, `0-9`, `-`, `_`
- **Padding**: None (no `=` characters)
- **Go Implementation**: `base64.RawURLEncoding`

### Key Derivation
**Workspace Keys**: Generated using `crypto/rand`
**Nonces**: Generated using `crypto/rand` (never reused)

---

## CLI Commands

### 1. Secret Creation

#### Command
```bash
initflow secret set <KEY> <VALUE> [OPTIONS]
```

#### Options
- `--workspace <id>` - Target workspace ID (required)
- `--description <text>` - Optional description for the secret
- `--force` - Overwrite existing secret without confirmation

#### Examples
```bash
# Basic secret creation
initflow secret set API_KEY "sk-1234567890abcdef" --workspace 42

# With description
initflow secret set DB_PASSWORD "super-secret-pass" --workspace 42 \
  --description "Production database password"

# Force overwrite
initflow secret set API_KEY "new-value" --workspace 42 --force
```

#### Cryptographic Workflow
1. **Validate Input**
   - Check workspace access permissions
   - Verify device authentication
   - Validate secret key format

2. **Retrieve Workspace Key**
   - Fetch device-wrapped workspace key from server
   - Decrypt using device's X25519 private key
   - Validate workspace key size (32 bytes)

3. **Encrypt Secret Value**
   ```go
   // Generate 24-byte nonce
   var nonce [24]byte
   rand.Read(nonce[:])
   
   // Encrypt with NaCl secretbox
   var key [32]byte
   copy(key[:], workspaceKey)
   ciphertext := secretbox.Seal(nil, []byte(value), &nonce, &key)
   ```

4. **Submit to Server**
   - Encode ciphertext and nonce as Base64 RawURL
   - Sign request with device Ed25519 private key
   - POST to `/api/v1/workspaces/{id}/secrets`

#### API Request Format
```json
{
  "key": "API_KEY",
  "encrypted_value": "base64-encoded-ciphertext",
  "nonce": "base64-encoded-24-byte-nonce"
}
```

### 2. Secret Retrieval

#### Command
```bash
initflow secret get <KEY> [OPTIONS]
```

#### Options
- `--workspace <id>` - Target workspace ID (required)
- `--output <format>` - Output format: `value`, `json`, `env` (default: `value`)
- `--copy` - Copy value to clipboard instead of printing

#### Examples
```bash
# Get secret value
initflow secret get API_KEY --workspace 42

# JSON output with metadata
initflow secret get API_KEY --workspace 42 --output json

# Environment variable format
initflow secret get API_KEY --workspace 42 --output env
```

#### Cryptographic Workflow
1. **Fetch Encrypted Secret**
   - GET `/api/v1/workspaces/{id}/secrets/{key}`
   - Receive encrypted_value and nonce from server

2. **Decrypt Secret Value**
   ```go
   // Decode from Base64
   ciphertext := base64.RawURLEncoding.DecodeString(encryptedValue)
   nonceBytes := base64.RawURLEncoding.DecodeString(nonce)
   
   // Convert to fixed arrays
   var nonce [24]byte
   var key [32]byte
   copy(nonce[:], nonceBytes)
   copy(key[:], workspaceKey)
   
   // Decrypt with NaCl secretbox
   plaintext, ok := secretbox.Open(nil, ciphertext, &nonce, &key)
   ```

#### Output Formats

**Value Format (Default)**
```
sk-1234567890abcdef
```

**JSON Format**
```json
{
  "key": "API_KEY",
  "value": "sk-1234567890abcdef",
  "version": 2,
  "workspace_id": 42,
  "updated_at": "2025-09-20T14:17:57Z",
  "created_by_device": "MacBook Pro"
}
```

**Environment Format**
```bash
export API_KEY="sk-1234567890abcdef"
```

### 3. Secret Listing

#### Command
```bash
initflow secret list [OPTIONS]
```

#### Options
- `--workspace <id>` - Target workspace ID (required)
- `--format <format>` - Output format: `table`, `json`, `simple` (default: `table`)

#### Examples
```bash
# Table format (default)
initflow secret list --workspace 42

# JSON format
initflow secret list --workspace 42 --format json

# Simple format (keys only)
initflow secret list --workspace 42 --format simple
```

#### Output Formats

**Table Format (Default)**
```
┌─────────────┬─────────┬─────────────────────┬──────────────┐
│ Key         │ Version │ Updated             │ Created By   │
├─────────────┼─────────┼─────────────────────┼──────────────┤
│ API_KEY     │ 2       │ 2h ago              │ MacBook Pro  │
│ DB_PASSWORD │ 1       │ 1d ago              │ Linux Server │
└─────────────┴─────────┴─────────────────────┴──────────────┘
```

**JSON Format**
```json
[
  {
    "key": "API_KEY",
    "version": 2,
    "updated_at": "2025-09-20T14:17:57Z",
    "created_by_device": "MacBook Pro"
  }
]
```

**Simple Format**
```
API_KEY
DB_PASSWORD
```

### 4. Secret Deletion

#### Command
```bash
initflow secret delete <KEY> [OPTIONS]
```

#### Options
- `--workspace <id>` - Target workspace ID (required)
- `--force` - Skip confirmation prompt

#### Examples
```bash
# Delete with confirmation
initflow secret delete API_KEY --workspace 42

# Force delete without confirmation
initflow secret delete API_KEY --workspace 42 --force
```

#### Workflow
1. **Confirmation Prompt** (unless `--force`)
2. **Submit Deletion Request**
   - DELETE `/api/v1/workspaces/{id}/secrets/{key}`
   - Server performs soft deletion (sets deleted_at timestamp)

---

## Security Considerations

### Threat Model Protection

| Threat | Mitigation |
|--------|------------|
| **Server Compromise** | Encrypted secrets useless without device private keys |
| **Network Interception** | All requests signed, sensitive data encrypted |
| **Device Theft** | Private keys protected by OS keychain + biometrics |
| **Replay Attacks** | Timestamp validation (5-minute window) |
| **Nonce Reuse** | Cryptographically random nonce generation |

### Implementation Security Checklist

#### Cryptographic Requirements
- ✅ Use `crypto/rand` for all random number generation
- ✅ Use NaCl secretbox for secret encryption
- ✅ Use Ed25519 for request signing
- ✅ Use X25519 for key wrapping
- ✅ Validate all cryptographic input sizes
- ✅ Use constant-time comparisons where applicable

#### Storage Security
- ✅ Store private keys in OS keychain
- ✅ Clear sensitive variables from memory after use
- ✅ Never log keys, secrets, or cryptographic material
- ✅ Validate server certificates in production

#### Error Handling
- ✅ Provide clear error messages without information leakage
- ✅ Fail securely (deny by default)
- ✅ Handle all edge cases gracefully

---

## API Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/v1/workspaces/{id}/secrets` | POST | Device | Create/update secret |
| `/api/v1/workspaces/{id}/secrets` | GET | Device | List secrets (metadata only) |
| `/api/v1/workspaces/{id}/secrets/{key}` | GET | Device | Get secret (encrypted) |
| `/api/v1/workspaces/{id}/secrets/{key}` | DELETE | Device | Delete secret |

### Authentication
All requests require device signature authentication:
- `Authorization: Device {device_id}`
- `X-Signature: {base64_encoded_ed25519_signature}`
- `X-Timestamp: {unix_timestamp}`

### Request Signing
**Signature Message Format** (body-agnostic):
```
{HTTP_METHOD}\n{REQUEST_PATH}{QUERY_STRING}\n{TIMESTAMP}
```

---

## Error Handling

### Common Error Scenarios

| Error | HTTP Status | CLI Response |
|-------|-------------|--------------|
| Invalid signature | 401 | `❌ Authentication failed. Re-register device` |
| Workspace not found | 404 | `❌ Workspace not found or not accessible` |
| Secret not found | 404 | `❌ Secret 'KEY' not found in workspace` |
| Access denied | 403 | `❌ Access denied to workspace` |
| Validation error | 422 | `❌ Invalid request parameters` |
| Decryption failure | - | `❌ Failed to decrypt secret: authentication failed` |

### Debugging Guidelines

1. **Verify Device Registration**
   ```bash
   initflow device list
   ```

2. **Check Workspace Access**
   ```bash
   initflow workspace list
   ```

3. **Validate Workspace Key**
   ```bash
   initflow workspace init <workspace-slug>
   ```

4. **Test Network Connectivity**
   ```bash
   initflow secret list --workspace <id>
   ```

---

## Migration and Compatibility

### Backward Compatibility
The CLI supports multiple encryption schemes for smooth migration:
- **Legacy**: ChaCha20-Poly1305 (12-byte nonces)
- **Current**: NaCl secretbox (24-byte nonces)
- **Future**: Quantum-resistant algorithms (when available)

### Migration Strategy
1. **Phase 1**: Deploy CLI with multi-scheme support
2. **Phase 2**: Migrate existing secrets to NaCl secretbox
3. **Phase 3**: Remove legacy encryption support

---

## Performance Considerations

### Encryption Performance
- **NaCl secretbox**: ~1GB/s on modern hardware
- **Key derivation**: Minimal overhead (<1ms)
- **Network latency**: Primary bottleneck for remote operations

### Optimization Strategies
- **Batch operations**: Group multiple secret operations
- **Local caching**: Cache workspace keys securely
- **Parallel requests**: Use concurrent API calls where safe

---

## Compliance and Auditing

### Standards Compliance
- **FIPS 140-2**: NaCl algorithms are FIPS-approved
- **SOC 2**: Zero-knowledge architecture supports compliance
- **GDPR**: No plaintext data stored on servers

### Audit Logging
Server maintains audit logs for:
- Secret creation/modification/deletion events
- Device authentication attempts
- Workspace access patterns
- API usage statistics

**Note**: Audit logs contain only metadata, never plaintext values.

---

## Development Guidelines

### Go Implementation Example
```go
package main

import (
    "crypto/rand"
    "golang.org/x/crypto/nacl/secretbox"
)

func encryptSecret(value string, workspaceKey []byte) ([]byte, []byte, error) {
    // Generate nonce
    var nonce [24]byte
    if _, err := rand.Read(nonce[:]); err != nil {
        return nil, nil, err
    }
    
    // Convert key
    var key [32]byte
    copy(key[:], workspaceKey)
    
    // Encrypt
    ciphertext := secretbox.Seal(nil, []byte(value), &nonce, &key)
    
    return ciphertext, nonce[:], nil
}

func decryptSecret(ciphertext, nonce, workspaceKey []byte) (string, error) {
    // Convert to arrays
    var nonceArray [24]byte
    var keyArray [32]byte
    copy(nonceArray[:], nonce)
    copy(keyArray[:], workspaceKey)
    
    // Decrypt
    plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
    if !ok {
        return "", errors.New("decryption failed")
    }
    
    return string(plaintext), nil
}
```

### Testing Requirements
- **Unit tests**: All cryptographic functions
- **Integration tests**: End-to-end secret lifecycle
- **Security tests**: Timing attacks, side-channel analysis
- **Compatibility tests**: Multiple encryption schemes

---

## Conclusion

This specification defines a secure, zero-knowledge secret management system that prioritizes user privacy and data security. The client-side encryption model ensures that even in the event of server compromise, user secrets remain protected.

For implementation questions or security concerns, refer to the development team or security audit documentation.

---

**Document Version History**
- v2.0 (2025-09-20): Updated to NaCl secretbox, clarified zero-knowledge architecture
- v1.0 (2025-09-13): Initial specification with ChaCha20-Poly1305

