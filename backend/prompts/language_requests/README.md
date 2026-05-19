# Language Requests

Use this folder to collect missing or provisional language keys during fast UI iteration.

Do not open or rewrite `backend/assets/language/languages.tsv` for every small UI edit. Record the key here first, then batch-update `languages.tsv` when the screen is release-ready or when translations are explicitly requested.

Each request should include:

- `key`
- Thai fallback
- English fallback
- screen/module
- reason or context
- caller file path

