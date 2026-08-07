---
kind: external_dependency
name: Firebase Authentication & Services
slug: firebase
category: external_dependency
category_hints:
    - vendor_identity
scope:
    - '**'
---

### Firebase
- **Role**: Google Firebase services for authentication, messaging, and cloud functions integration.
- **Integration**: firebase.google.com/go/v4 SDK with Firestore and Storage clients.
- **Configuration**: FIREBASE_PROJECT_ID environment variable; supports both development and production Firebase projects.
- **Usage Pattern**: Used for mobile app authentication, push notifications, and cloud function triggers.
- **Environment**: Development project ID 'test06-a1240' configured in docker-compose for local development.