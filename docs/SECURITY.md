# Security Architecture

This document provides a detailed technical description of Initiat's security architecture, focusing on the zero-knowledge approach, encryption mechanisms, and cryptographic protocols used to ensure maximum security for secret management.

## Table of Contents

- [Zero-Knowledge Architecture](#zero-knowledge-architecture)
- [Cryptographic Foundations](#cryptographic-foundations)
- [Key Management](#key-management)
- [Encryption Protocols](#encryption-protocols)
- [Authentication & Authorization](#authentication--authorization)
- [Secure Storage](#secure-storage)
- [Network Security](#network-security)
- [Threat Model](#threat-model)
- [Security Guarantees](#security-guarantees)

## Zero-Knowledge Architecture

Initiat implements a true zero-knowledge architecture where the server has no access to plaintext secrets or workspace keys. This ensures that even if the server is compromised, your secrets remain secure.

### Core Principles

1. **Client-Side Encryption**: All secrets are encrypted on the client before transmission
2. **Server Blindness**: The server never sees plaintext secrets or workspace keys
3. **Key Isolation**: Workspace keys are generated client-side and never transmitted in plaintext
4. **Forward Secrecy**: Compromising one device doesn't affect other devices
5. **Audit Trail**: All operations are cryptographically signed and logged

### Zero-Knowledge Guarantees

- **Secrets**: Server cannot decrypt your secrets, even with full database access
- **Workspace Keys**: Server cannot access workspace keys, even with admin privileges
- **Device Keys**: Private keys never leave the client device
- **Authentication**: Passwords are never stored, only used for initial authentication

## Cryptographic Foundations

### Cryptographic Primitives

Initiat uses industry-standard cryptographic primitives with proven security:

| Component | Algorithm | Key Size | Purpose |
|-----------|-----------|---------|---------|
| **Signing** | Ed25519 | 256-bit | Device authentication, request signing |
| **Encryption** | X25519 | 256-bit | Key exchange, workspace key wrapping |
| **Secret Encryption** | XSalsa20Poly1305 | 256-bit | Secret value encryption |
| **Key Wrapping** | ChaCha20Poly1305 | 256-bit | Workspace key encryption |
| **Key Derivation** | HKDF-SHA256 | 256-bit | Key derivation from shared secrets |
| **Random Generation** | OS CSPRNG | 256-bit | All cryptographic randomness |

### Algorithm Selection Rationale

- **Ed25519**: Modern, high-performance elliptic curve signature scheme
- **X25519**: Efficient, secure key exchange protocol
- **XSalsa20Poly1305**: Authenticated encryption with additional data (AEAD)
- **ChaCha20Poly1305**: High-performance AEAD cipher
- **HKDF-SHA256**: Secure key derivation function

## Key Management

### Device Key Generation

Each device generates two keypairs during registration. The implementation can be found in `internal/crypto/crypto.go` in the `GenerateEd25519Keypair` and `GenerateX25519Keypair` functions.

**Process:**
1. Generate Ed25519 signing keypair for device authentication
2. Generate X25519 encryption keypair for key exchange
3. Store private keys securely in OS keychain

### Workspace Key Generation

Workspace keys are generated client-side using cryptographically secure random number generation. The implementation can be found in `internal/crypto/crypto.go`.

**Process:**
1. Generate 256-bit (32-byte) random workspace key
2. Use OS CSPRNG for cryptographic randomness
3. Key is generated locally and never transmitted in plaintext

### Key Storage

All private keys are stored securely in the operating system's credential management system:

- **macOS**: Keychain Services with `kSecClassGenericPassword`
- **Windows**: Windows Credential Manager with `CRED_TYPE_GENERIC`
- **Linux**: Secret Service (GNOME Keyring, KDE Wallet, etc.)

## Encryption Protocols

### Secret Value Encryption

Secrets are encrypted using XSalsa20Poly1305 (NaCl secretbox) with workspace keys. The implementation can be found in `internal/crypto/crypto.go` in the `EncryptSecretValue` function.

**Process:**
1. Generate a random 24-byte nonce
2. Use the workspace key as the encryption key
3. Encrypt the secret value using authenticated encryption
4. Return the ciphertext and nonce

**Security Properties:**
- **Authenticated Encryption**: Prevents tampering and ensures authenticity
- **Nonce Randomness**: Each encryption uses a unique random nonce
- **Key Isolation**: Each workspace has its own encryption key

### Workspace Key Wrapping

Workspace keys are wrapped using X25519 key exchange and ChaCha20Poly1305. The implementation can be found in `internal/crypto/crypto.go` in the `WrapWorkspaceKey` function.

**Process:**
1. Generate an ephemeral X25519 keypair
2. Compute shared secret using X25519 key agreement
3. Derive encryption key using HKDF-SHA256
4. Encrypt workspace key with ChaCha20Poly1305
5. Package ephemeral public key, nonce, and ciphertext
6. Encode the result in base64url format

**Security Properties:**
- **Forward Secrecy**: Each wrapping uses a unique ephemeral keypair
- **Key Agreement**: Only the intended device can unwrap the key
- **Authenticated Encryption**: Prevents tampering during transmission

### Key Unwrapping

The unwrapping process reverses the wrapping operation. The implementation can be found in `internal/crypto/crypto.go` in the `UnwrapWorkspaceKey` function.

**Process:**
1. Decode the base64url-encoded wrapped key
2. Extract ephemeral public key, nonce, and ciphertext
3. Compute shared secret using X25519 key agreement
4. Derive decryption key using HKDF-SHA256
5. Decrypt workspace key with ChaCha20Poly1305
6. Return the decrypted workspace key

## Authentication & Authorization

### Device Authentication

All API requests are authenticated using Ed25519 digital signatures following RFC 8032. The signature is computed over a canonicalized request payload and included in the `X-Initiat-Signature` header.

**Process:**
1. Create canonical request representation
2. Sign request data with device's Ed25519 private key
3. Encode signature in base64url format
4. Include signature in request headers

### Request Signing Protocol

1. **Timestamp**: Include current timestamp to prevent replay attacks
2. **Request Hash**: Hash request body with SHA-256
3. **Signature Data**: Combine timestamp, method, path, and body hash
4. **Sign**: Create Ed25519 signature of combined data
5. **Headers**: Include signature and timestamp in request headers

### Authorization Model

- **Device Registration**: Requires valid authentication token
- **Workspace Access**: Requires device approval by workspace admin
- **Secret Operations**: Requires workspace key access
- **Admin Operations**: Requires workspace admin role

## Secure Storage

### OS Keychain Integration

The CLI integrates with the operating system's secure credential storage. The implementation can be found in `internal/storage/` with platform-specific implementations.

#### macOS Keychain Services
Private keys are stored using `kSecClassGenericPassword` with device-only access (`kSecAttrAccessibleWhenUnlockedThisDeviceOnly`). Keys are identified by service name "initiat-cli" and account identifiers.

#### Windows Credential Manager
Private keys are stored using `CRED_TYPE_GENERIC` with local machine persistence. Credentials are identified by target name "initiat-cli/device-signing-key".

#### Linux Secret Service
Private keys are stored using the Secret Service API (D-Bus) with service identification "initiat-cli". The implementation supports GNOME Keyring, KDE Wallet, and other Secret Service providers.

### Storage Security Properties

- **Hardware Security**: Keys stored in secure hardware when available
- **Access Control**: Keys only accessible to the user account
- **Encryption at Rest**: Keys encrypted by the OS credential system
- **No Network Access**: Private keys never transmitted over network

## Network Security

### Transport Layer Security

All network communication uses HTTPS through Go's standard `net/http` client. The implementation can be found in `internal/client/client.go` and `internal/httputil/httputil.go`.

**Security Features:**
- **HTTPS Only**: All API requests use HTTPS
- **Request Signing**: All authenticated requests signed with Ed25519
- **Timestamp Validation**: Prevents replay attacks

### Request Security

The CLI implements request signing for authenticated API communications. The implementation can be found in `internal/httputil/httputil.go` in the `SignRequest` function.

**Request Signing Process:**
1. Create canonical request representation (method, path, timestamp)
2. Sign request data with device's Ed25519 private key
3. Include signature and timestamp in request headers

**Security Headers:**
- `Authorization: Device <device-id>` - Device identification
- `X-Signature: <ed25519-signature>` - Request signature
- `X-Timestamp: <unix-timestamp>` - Timestamp for replay protection

## Threat Model

### Adversarial Capabilities

Initiat's security model protects against:

#### Server Compromise
- **Database Access**: Attacker gains full database access
- **Code Execution**: Attacker can execute arbitrary code on server
- **Network Interception**: Attacker can intercept all network traffic
- **Log Access**: Attacker can read all server logs

#### Network Attacks
- **Man-in-the-Middle**: Attacker intercepts network traffic
- **DNS Spoofing**: Attacker redirects DNS queries
- **Certificate Attacks**: Attacker compromises certificate authority

#### Client Device Compromise
- **Key Extraction**: Attacker extracts private keys from device
- **Memory Dumps**: Attacker gains access to process memory
- **Keychain Access**: Attacker accesses OS credential storage

### Security Guarantees

#### Server Compromise Protection
- **Zero-Knowledge**: Server cannot decrypt secrets even with full access
- **Key Isolation**: Workspace keys never stored on server in plaintext
- **Forward Secrecy**: Compromising server doesn't affect historical data

#### Network Attack Protection
- **HTTPS**: All traffic encrypted using standard TLS
- **Request Signing**: Prevents request tampering and replay attacks
- **Timestamp Validation**: Prevents replay attacks

#### Client Compromise Protection
- **Key Isolation**: Compromising one device doesn't affect others
- **Secure Storage**: Keys protected by OS credential system
- **Memory Protection**: Keys cleared from memory after use

## Security Guarantees

### Cryptographic Guarantees

1. **Confidentiality**: Secrets remain confidential even if server is compromised
2. **Integrity**: All data is cryptographically authenticated
3. **Authenticity**: All requests are cryptographically signed
4. **Non-repudiation**: All operations are cryptographically logged
5. **Forward Secrecy**: Compromising current keys doesn't affect past data

### Operational Guarantees

1. **Zero-Knowledge**: Server has no access to plaintext secrets
2. **Key Isolation**: Each workspace has independent encryption keys
3. **Device Independence**: Compromising one device doesn't affect others
4. **Audit Trail**: All operations are cryptographically signed and logged
5. **Secure Deletion**: Keys can be securely revoked and deleted

### Compliance & Standards

Initiat uses industry-standard cryptographic algorithms and security practices:

- **Cryptographic Standards**: Uses Ed25519 signatures as defined in [RFC 8032](https://tools.ietf.org/html/rfc8032) and standard cryptographic libraries
- **Security Best Practices**: Implements secure coding practices including input validation and secure key storage
- **Data Protection**: Implements zero-knowledge architecture for data privacy

**Note**: Specific compliance certifications (FIPS, SOC 2, etc.) are not currently implemented but will be pursued in future versions.

## Security Best Practices

### For Users

1. **Device Security**: Keep devices updated and use strong authentication
2. **Key Management**: Never share device credentials or workspace keys
3. **Access Control**: Regularly audit device access and remove unused devices
4. **Secret Rotation**: Regularly rotate secrets and update versions
5. **Backup**: Ensure secure backup of critical workspace keys

### For Administrators

1. **Device Approval**: Carefully review device approval requests
2. **Access Auditing**: Regularly review device access and permissions
3. **Incident Response**: Have procedures for device compromise
4. **Key Rotation**: Implement regular key rotation policies
5. **Monitoring**: Monitor for unusual access patterns

### For Developers

1. **Secure Coding**: Follow secure coding practices
2. **Dependency Management**: Keep dependencies updated
3. **Security Testing**: Regular security testing and audits
4. **Incident Response**: Have procedures for security incidents
5. **Documentation**: Maintain security documentation

## Security Audit

### Cryptographic Review

Initiat's cryptographic implementation uses industry-standard algorithms and libraries:

- **Algorithm Selection**: Uses well-reviewed cryptographic algorithms (Ed25519, X25519, ChaCha20Poly1305)
- **Key Management**: Implements proper key generation and storage
- **Protocol Design**: Follows established cryptographic protocols
- **Implementation**: Uses Go's standard library and golang.org/x/crypto

### Third-Party Security

- **Dependencies**: Minimal, well-maintained cryptographic dependencies
- **Libraries**: Uses Go's standard library and golang.org/x/crypto
- **Vulnerability Scanning**: Uses govulncheck for vulnerability detection
- **Security Updates**: Regular updates of dependencies

## Conclusion

Initiat's security architecture provides enterprise-grade security through:

- **Zero-Knowledge Design**: Server cannot access plaintext secrets
- **Strong Cryptography**: Industry-standard cryptographic algorithms
- **Secure Key Management**: Proper key generation, storage, and rotation
- **Defense in Depth**: Multiple layers of security controls
- **Audit Trail**: Comprehensive logging and monitoring

This architecture ensures that your secrets remain secure even in the face of sophisticated attacks, providing the confidence needed for enterprise adoption.
