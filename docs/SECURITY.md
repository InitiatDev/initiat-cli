# Security

**Offline-first: core workflows do not require sending code or secrets to any server.**

This document explains Initiat's security model. The main point: **the core product runs entirely on your machine using YAML in your repo. No account, no server, no transmission of your code or secrets.**

## Quick Summary (Core / Offline)

**What you need to know about the default experience:**

- **No server required**: Setup, validation, and docs generation run locally. Nothing is sent to Initiat.
- **No account required**: You do not need to sign up or log in to use the CLI for in-repo workflows.
- **Secrets stay with you**: The core product does not store or transmit secrets. Use your own secret store (e.g. env files, 1Password CLI, Vault) and wire them into setup via provider-agnostic config.
- **Reproducible and auditable**: All behavior is defined by YAML in the repo. You can review and version every step.

**In practice (offline):**
- You define setup in `.initiat/setup.yml` and optionally env/docs in other `.initiat/` files.
- You run `initiat setup validate`, `initiat setup run`, `initiat docs generate` locally.
- No network call to Initiat is made for these operations. Your code and secrets never leave your environment.

---

## Optional Cloud Add-on

If your team opts in to hosted features (e.g. shared secret storage, device approval), the following applies only to that **optional** usage.

### Zero-Knowledge Secret Management (Optional)

When you use Initiat's hosted secret storage:

- **Secrets are encrypted on your device** before being sent to our servers.
- **We cannot decrypt your secrets** — even with full server access.
- **Each project has its own encryption key** that never leaves your device unencrypted.
- **Your private keys are stored** in your OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service).
- **Each device has its own keys** — compromising one device doesn't affect others.

**In practice (cloud add-on):**
- You create an account and register devices only if you choose to use hosted secrets.
- When you set a secret via the cloud add-on, it's encrypted on your computer and only ciphertext is sent.
- When you retrieve it, your device decrypts it locally. We never see plaintext.

### How Optional Hosted Secrets Work

1. **You set a secret**: `initiat secret set API_KEY --value "sk-123..."` (requires cloud/account).
2. **Your device encrypts it** using a project-specific key.
3. **The encrypted secret is sent** to our servers (we can't read it).
4. **When you retrieve it**: Your device downloads the ciphertext and decrypts it locally.
5. **The secret is used** in your environment but never logged in plaintext.

### Optional Cloud: Technical Details

<details>
<summary><strong>Secret Encryption Process (Click to expand)</strong></summary>

When you set a secret via the cloud add-on:

1. **Project Key Retrieval**: Your device retrieves the project's encryption key from secure storage.
2. **Nonce Generation**: A random 24-byte nonce is generated.
3. **Encryption**: The secret value is encrypted using XSalsa20Poly1305 (authenticated encryption).
4. **Storage**: Only the encrypted secret (ciphertext + nonce) is sent to our servers.

**Cryptographic Details:**
- **Algorithm**: XSalsa20Poly1305 (NaCl secretbox)
- **Key Size**: 256-bit (32 bytes)
- **Nonce Size**: 24 bytes (random per encryption)
- **Security**: Authenticated encryption; server cannot decrypt even with full database access.

</details>

<details>
<summary><strong>Project Key Management (Click to expand)</strong></summary>

Each project has its own 256-bit encryption key, generated client-side and never transmitted in plaintext.

**Key Sharing (Device Invitation):**
1. **Key Wrapping**: The project key is encrypted for the new device.
2. **Key Exchange**: X25519 creates a shared secret; ChaCha20Poly1305 encrypts the key.
3. **Transmission**: Only the wrapped key is sent; server cannot unwrap it.
4. **Unwrapping**: The new device unwraps using its private key.

**Cryptographic Details:**
- **Key Exchange**: X25519
- **Key Derivation**: HKDF-SHA256
- **Encryption**: ChaCha20Poly1305

</details>

<details>
<summary><strong>Device Authentication (Click to expand)</strong></summary>

Every API request for the cloud add-on is signed with your device's Ed25519 private key. Servers verify the signature and use timestamps for replay protection.

**Algorithm**: Ed25519 (RFC 8032). What gets signed: HTTP method, path, body hash, timestamp.

</details>

---

## Key Storage (Cloud Add-on Only)

When using the cloud add-on, private keys are stored in your OS credential store:

- **macOS**: Keychain Services
- **Windows**: Windows Credential Manager
- **Linux**: Secret Service (GNOME Keyring, KDE Wallet, etc.)

Keys are encrypted at rest by the OS and only accessible to your user account.

---

## Threat Model

### Core (Offline)

- **No server, no transmission**: There is nothing to compromise on our side for core workflows. We don't receive your code or secrets.
- **Local execution**: All risk is local (your machine, your repo, your secret provider). Use standard practices: restrict repo access, use a trusted secret store, keep the CLI and OS updated.

### Optional Cloud Add-on

- **Server compromise**: We cannot decrypt your secrets; they are encrypted with keys we don't have.
- **Network attacks**: HTTPS and request signing protect in transit and from tampering.
- **Device compromise**: If your device is compromised, an attacker can use your keys. Revoke the device and use other devices; forward secrecy limits blast radius.

---

## Security Guarantees

**Core (offline):**
- No Initiat server sees your code or secrets when you use setup, validate, or docs generation.
- Behavior is fully defined by repo YAML; you can audit and version it.

**Optional cloud add-on:**
- Confidentiality: We cannot decrypt your secrets.
- Integrity: Data is cryptographically authenticated.
- Authenticity: Requests are signed by your device.
- Forward secrecy: Compromising one device doesn't expose others.

**Algorithms (cloud add-on):** Ed25519, X25519, ChaCha20Poly1305, XSalsa20Poly1305, HKDF-SHA256. Implementations use Go standard library and golang.org/x/crypto.

---

## Best Practices

### For Everyone (Offline or Cloud)

- Keep your OS and CLI updated.
- Use strong authentication on your machine.
- Prefer the offline workflow when you don't need shared hosted secrets.
- For secrets, use a dedicated store (e.g. 1Password, Vault, gitignored `.env`) and wire it via Initiat's provider-agnostic env config; don't commit secrets.

### For Cloud Add-on Users

- Review and approve devices carefully.
- Rotate secrets periodically.
- Revoke unused devices.
- Educate the team on what is stored where (we only store encrypted blobs; keys stay on devices).

---

## Questions?

**Q: Do I have to send my code or secrets to Initiat to use the CLI?**  
A: No. Core workflows (setup, validate, docs) run entirely locally. No account or server is required. Secrets are out of scope for core; use your own secret store.

**Q: If I use the optional cloud secret storage, can Initiat employees see my secrets?**  
A: No. Secrets are encrypted on your device before transmission; we don't have the keys to decrypt them.

**Q: What if I only want in-repo setup and docs?**  
A: Use `initiat init`, `initiat setup validate`, `initiat setup run`, and `initiat docs generate`. No signup, no cloud, no secret storage on our side.

**Q: How do I know the core is secure?**  
A: Core doesn't send your data anywhere. You run the binary and YAML locally; you can audit the code and the repo config.

---

**Implementation details for the optional cloud add-on:** `internal/crypto/`, `internal/storage/`, `internal/httputil/`, `internal/auth/`.
