# Security & Secret Management

**How Initiat protects your secrets with zero-knowledge architecture**

This document explains how Initiat handles secrets securely. For average developers, the key takeaway is: **your secrets are encrypted on your device before they ever reach our servers, and we can never decrypt them**. For security experts, we've included detailed technical sections you can expand.

## Quick Summary

**What you need to know:**

- 🔐 **Secrets are encrypted on your device** before being sent to our servers
- 🚫 **We cannot decrypt your secrets** - even with full server access
- 🔑 **Each project has its own encryption key** that never leaves your device unencrypted
- ✅ **Your private keys are stored** in your OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- 🔄 **Each device has its own keys** - compromising one device doesn't affect others

**In practice:**
- When you set a secret, it's encrypted on your computer using a project key
- The encrypted secret is sent to our servers (we can't read it)
- When you retrieve a secret, it's decrypted on your computer
- All of this happens automatically - you just use `initiat secret set` and `initiat secret get`

---

## How Secrets Are Handled

### The Simple Version

1. **You set a secret**: `initiat secret set API_KEY --value "sk-123..."`
2. **Your device encrypts it** using a project-specific key
3. **The encrypted secret is sent** to our servers (we can't read it)
4. **When you retrieve it**: Your device downloads the encrypted secret and decrypts it locally
5. **The secret is used** in your environment (e.g., in setup scripts) but never logged

### The Technical Version

<details>
<summary><strong>Secret Encryption Process (Click to expand)</strong></summary>

When you set a secret, here's what happens under the hood:

1. **Project Key Retrieval**: Your device retrieves the project's encryption key from secure storage
2. **Nonce Generation**: A random 24-byte nonce is generated for this encryption
3. **Encryption**: The secret value is encrypted using XSalsa20Poly1305 (authenticated encryption)
4. **Storage**: The encrypted secret (ciphertext + nonce) is sent to our servers

**Cryptographic Details:**
- **Algorithm**: XSalsa20Poly1305 (NaCl secretbox)
- **Key Size**: 256-bit (32 bytes)
- **Nonce Size**: 24 bytes (randomly generated for each encryption)
- **Security Properties**: Authenticated encryption prevents tampering and ensures authenticity

**Why this matters:**
- Each encryption uses a unique random nonce, so the same secret encrypted twice produces different ciphertext
- The server only sees encrypted data - it cannot decrypt it even with full database access
- The encryption is authenticated, meaning any tampering would be detected

</details>

<details>
<summary><strong>Project Key Management (Click to expand)</strong></summary>

Each project has its own 256-bit encryption key that's generated client-side and never transmitted in plaintext.

**Key Generation:**
- Generated on your device using OS cryptographically secure random number generation
- Never sent to our servers in plaintext
- Stored securely in your OS keychain

**Key Sharing (Device Invitation):**
When you invite a new device to a project:

1. **Key Wrapping**: The project key is "wrapped" (encrypted) specifically for the new device
2. **Ephemeral Key**: A temporary keypair is generated for this wrapping operation
3. **Key Exchange**: Uses X25519 key exchange to create a shared secret
4. **Encryption**: The project key is encrypted using ChaCha20Poly1305
5. **Transmission**: Only the wrapped (encrypted) key is sent to our servers
6. **Unwrapping**: The new device can unwrap the key using its private key

**Security Properties:**
- **Forward Secrecy**: Each wrapping uses a unique ephemeral keypair
- **Key Isolation**: Only the intended device can unwrap the key
- **Server Blindness**: The server never sees the plaintext project key

**Cryptographic Details:**
- **Key Exchange**: X25519 (Elliptic Curve Diffie-Hellman)
- **Key Derivation**: HKDF-SHA256
- **Encryption**: ChaCha20Poly1305 (authenticated encryption)
- **Key Size**: 256-bit (32 bytes)

</details>

<details>
<summary><strong>Device Authentication (Click to expand)</strong></summary>

Every API request is cryptographically signed to prove it came from your device.

**How it works:**
1. **Request Signing**: Your device signs each request with its Ed25519 private key
2. **Signature Verification**: Our servers verify the signature using your device's public key
3. **Replay Protection**: Timestamps prevent old requests from being replayed

**Cryptographic Details:**
- **Algorithm**: Ed25519 (RFC 8032)
- **Key Size**: 256-bit private key, 256-bit public key
- **Signature Size**: 512 bits (64 bytes)
- **Performance**: Very fast signing and verification

**What gets signed:**
- HTTP method (GET, POST, etc.)
- Request path
- Request body hash (SHA-256)
- Timestamp

**Security Benefits:**
- **Authentication**: Proves the request came from your device
- **Integrity**: Prevents request tampering
- **Non-repudiation**: Cryptographically proves who made the request

</details>

---

## Zero-Knowledge Architecture

**What "zero-knowledge" means:**

We designed Initiat so that even if our servers are completely compromised, your secrets remain secure. Here's how:

### Server Compromise Protection

**Even if an attacker:**
- Gains full database access
- Can execute code on our servers
- Intercepts all network traffic
- Reads all server logs

**They still cannot:**
- Decrypt your secrets (they're encrypted with keys we don't have)
- Access project keys (they're wrapped and we can't unwrap them)
- Impersonate your device (they don't have your private keys)

### How It Works

1. **Client-Side Encryption**: All encryption happens on your device before transmission
2. **Key Isolation**: Project keys are generated client-side and never sent in plaintext
3. **Device Keys**: Your device's private keys never leave your device
4. **Forward Secrecy**: Compromising one device doesn't affect others

---

## Key Storage

Your private keys are stored securely using your operating system's built-in credential management:

- **macOS**: Keychain Services
- **Windows**: Windows Credential Manager  
- **Linux**: Secret Service (GNOME Keyring, KDE Wallet, etc.)

**What this means:**
- Keys are encrypted at rest by your OS
- Keys are only accessible to your user account
- Keys may be stored in secure hardware (e.g., Secure Enclave on macOS)
- Keys never leave your device

<details>
<summary><strong>Storage Implementation Details (Click to expand)</strong></summary>

**macOS Keychain Services:**
- Uses `kSecClassGenericPassword` with `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`
- Keys are device-only (not synced to iCloud)
- Protected by user authentication
- May use Secure Enclave on supported devices

**Windows Credential Manager:**
- Uses `CRED_TYPE_GENERIC` with local machine persistence
- Protected by Windows user account security
- Credentials identified by service name and account

**Linux Secret Service:**
- Uses D-Bus Secret Service API
- Supports GNOME Keyring, KDE Wallet, and other providers
- Protected by user session authentication
- Keys identified by service name "initiat-cli"

</details>

---

## Network Security

All communication with our servers uses HTTPS (TLS 1.2+). Additionally:

- **Request Signing**: Every authenticated request is cryptographically signed
- **Timestamp Validation**: Prevents replay attacks
- **Certificate Validation**: Standard TLS certificate validation

**What this protects against:**
- Man-in-the-middle attacks
- Network interception
- Request tampering
- Replay attacks

---

## Threat Model

**What we protect against:**

### Server Compromise
✅ **Protected**: Even with full server access, secrets cannot be decrypted

### Network Attacks
✅ **Protected**: HTTPS + request signing prevents interception and tampering

### Device Compromise
⚠️ **Partial Protection**: If your device is compromised, the attacker can access your keys. However:
- Compromising one device doesn't affect other devices
- Keys are protected by OS security
- You can revoke compromised devices

**Best Practices:**
- Keep your devices updated
- Use strong authentication (passwords, biometrics)
- Revoke access for lost or compromised devices
- Regularly audit device access

---

## Security Guarantees

**What we guarantee:**

1. **Confidentiality**: Your secrets remain confidential even if our servers are compromised
2. **Integrity**: All data is cryptographically authenticated to prevent tampering
3. **Authenticity**: All requests are cryptographically signed to prove they came from your device
4. **Non-repudiation**: All operations are cryptographically logged
5. **Forward Secrecy**: Compromising current keys doesn't affect past encrypted data

**What we use:**

- **Industry-standard algorithms**: Ed25519, X25519, ChaCha20Poly1305, XSalsa20Poly1305
- **Proven cryptographic libraries**: Go standard library and golang.org/x/crypto
- **Secure key storage**: OS credential management systems
- **Best practices**: Proper key generation, secure random number generation, authenticated encryption

---

## For Security Experts

<details>
<summary><strong>Cryptographic Primitives (Click to expand)</strong></summary>

| Component | Algorithm | Key Size | Purpose |
|-----------|-----------|---------|---------|
| **Signing** | Ed25519 | 256-bit | Device authentication, request signing |
| **Key Exchange** | X25519 | 256-bit | Project key wrapping |
| **Secret Encryption** | XSalsa20Poly1305 | 256-bit | Secret value encryption |
| **Key Wrapping** | ChaCha20Poly1305 | 256-bit | Project key encryption |
| **Key Derivation** | HKDF-SHA256 | 256-bit | Key derivation from shared secrets |
| **Random Generation** | OS CSPRNG | 256-bit | All cryptographic randomness |

**Algorithm Selection Rationale:**
- **Ed25519**: Modern, high-performance elliptic curve signature scheme (RFC 8032)
- **X25519**: Efficient, secure key exchange protocol
- **XSalsa20Poly1305**: Authenticated encryption with additional data (AEAD)
- **ChaCha20Poly1305**: High-performance AEAD cipher
- **HKDF-SHA256**: Secure key derivation function (RFC 5869)

</details>

<details>
<summary><strong>Implementation Details (Click to expand)</strong></summary>

**Key Files:**
- `internal/crypto/crypto.go`: Cryptographic operations (encryption, key generation, signing)
- `internal/storage/storage.go`: OS keychain integration
- `internal/httputil/httputil.go`: Request signing
- `internal/auth/auth.go`: Authentication and device registration

**Key Generation:**
- Uses `crypto/rand` for all random number generation
- Ed25519 keypairs generated using `golang.org/x/crypto/ed25519`
- X25519 keypairs generated using `golang.org/x/crypto/curve25519`

**Encryption:**
- Secret encryption: `golang.org/x/crypto/nacl/secretbox`
- Key wrapping: `golang.org/x/crypto/chacha20poly1305`
- Key derivation: `golang.org/x/crypto/hkdf`

**Storage:**
- macOS: `github.com/keybase/go-keychain`
- Windows: `golang.org/x/sys/windows`
- Linux: `github.com/godbus/dbus/v5` (Secret Service)

</details>

<details>
<summary><strong>Security Audit & Compliance (Click to expand)</strong></summary>

**Current Status:**
- Uses industry-standard cryptographic algorithms
- Implements secure coding practices
- Regular dependency updates and vulnerability scanning
- Code review process for security-sensitive changes

**Future Plans:**
- Third-party security audits
- Compliance certifications (FIPS, SOC 2, etc.)
- Bug bounty program
- Regular penetration testing

**Dependencies:**
- Minimal, well-maintained cryptographic dependencies
- Regular updates via `go get -u` and vulnerability scanning
- Uses `govulncheck` for vulnerability detection

</details>

---

## Best Practices

### For Users

1. **Keep devices updated** - Security updates protect against vulnerabilities
2. **Use strong authentication** - Protect your device with passwords/biometrics
3. **Review device access** - Regularly check which devices have access
4. **Revoke unused devices** - Remove access for devices you no longer use
5. **Rotate secrets regularly** - Update secrets periodically for security

### For Administrators

1. **Review device approvals** - Carefully approve device access requests
2. **Audit access regularly** - Review who has access to what projects
3. **Have incident response** - Plan for device compromise scenarios
4. **Monitor access patterns** - Watch for unusual activity
5. **Educate your team** - Ensure everyone understands security practices

---

## Questions?

**Common Questions:**

**Q: Can Initiat employees see my secrets?**  
A: No. Secrets are encrypted on your device before transmission, and we don't have the keys to decrypt them.

**Q: What happens if I lose my device?**  
A: You can revoke the device's access. The device won't be able to access secrets anymore, but other devices remain unaffected.

**Q: Can I recover secrets if I lose all my devices?**  
A: This depends on your backup strategy. Project keys are stored in your OS keychain, so if you have keychain backups, you can recover. Otherwise, you may need to re-encrypt secrets with a new project key.

**Q: How do I know it's secure?**  
A: We use industry-standard cryptographic algorithms that are widely reviewed and trusted. The zero-knowledge architecture means even we can't decrypt your secrets, which is the strongest guarantee possible.

---

**For more technical details, see the implementation in `internal/crypto/` and `internal/storage/`.**
