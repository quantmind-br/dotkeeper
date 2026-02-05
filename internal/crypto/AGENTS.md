# CRYPTO PACKAGE KNOWLEDGE BASE

**Scope:** `internal/crypto/`

## OVERVIEW

Cryptographic primitives: AES-256-GCM + Argon2id. 6 files, ~800 lines. **Security-critical.**

## STRUCTURE

```
crypto/
├── aes.go       # AES-256-GCM encryption/decryption
├── kdf.go       # Argon2id key derivation
├── types.go     # Constants & metadata types
├── aes_test.go
├── kdf_test.go
└── types_test.go
```

## CIPHERTEXT FORMAT

```
[version(1 byte)][salt(16 bytes)][nonce(12 bytes)][ciphertext...][tag(16 bytes)]
```

## CONSTANTS

```go
// types.go - NEVER CHANGE THESE
const (
    Version        = 1
    SaltSize       = 16
    NonceSize      = 12
    KeySize        = 32
    Argon2Time     = 3
    Argon2Memory   = 64 * 1024  // 64 MB
    Argon2Threads  = 4
)
```

## API

```go
// Encrypt returns full ciphertext including header
func Encrypt(plaintext []byte, password string) ([]byte, *Metadata, error)

// Decrypt extracts header, derives key, decrypts
func Decrypt(ciphertext []byte, password string) ([]byte, error)

// DeriveKey uses Argon2id
func DeriveKey(password string, salt []byte) []byte
```

## CONVENTIONS

- **Random**: `crypto/rand` for salt and nonce
- **Timing**: Argon2id params tuned for ~100ms on modern CPUs
- **Validation**: Check version byte, verify tag (GCM built-in)

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Change KDF params | `types.go` | **NEVER** - would break existing backups |
| Add keyfile support | `kdf.go` | Combine password + keyfile |
| Streaming encryption | `aes.go` | **TODO** - currently loads all in memory |
| Cipher upgrade path | `aes.go` | Version byte reserved for v2 |

## ANTI-PATTERNS

- **NEVER** change Argon2id parameters (breaks backward compatibility)
- **NEVER** use non-random salts/nonces
- **NEVER** skip version validation
- **NEVER** roll your own crypto (use stdlib + golang.org/x/crypto)
- **Don't** use //nolint without security justification

## SECURITY NOTES

- Metadata (`.meta.json`) is NOT encrypted - contains salt + KDF params only
- Password never stored - derived key used for encryption
- Memory clearing: Go doesn't guarantee zeroing, but minimizes exposure
