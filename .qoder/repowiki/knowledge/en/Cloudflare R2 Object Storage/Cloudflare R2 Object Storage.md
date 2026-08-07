---
kind: external_dependency
name: Cloudflare R2 Object Storage
slug: cloudflare-r2
category: external_dependency
category_hints:
    - vendor_identity
    - auth_protocol
scope:
    - '**'
---

### Cloudflare R2
- **Role**: S3-compatible object storage for images, files, and binary assets in BC Ai Account system.
- **Integration**: AWS SDK v2 S3 client configured with R2-specific endpoints and credentials.
- **Authentication**: Uses R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY environment variables with optional custom endpoint support.
- **Usage Pattern**: All uploaded images/files stored as actual files (not base64); references stored in MongoDB documents.
- **Configuration**: Supports path-style URLs via R2_FORCE_PATH_STYLE flag for local MinIO compatibility.