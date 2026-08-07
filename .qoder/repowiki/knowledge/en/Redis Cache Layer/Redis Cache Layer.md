---
kind: external_dependency
name: Redis Cache Layer
slug: redis
category: external_dependency
category_hints:
    - vendor_identity
scope:
    - '**'
---

### Redis
- **Role**: Caching layer for authentication tokens, session management, and application-level caching.
- **Integration**: go-redis/v8 client with connection URI and password configuration.
- **Usage Pattern**: Used for auth token caching with TTL-based expiration (24 hours access tokens, 30 days refresh tokens).
- **Deployment**: Runs in Docker container on port 6379, accessible via localhost from mainapi container.
- **Configuration**: REDIS_CACHE_URI and REDIS_CACHE_PASSWORD environment variables.