# Fly.io Security Configuration

## Required Environment Variables

Set these secrets before deploying:

```bash
# Required
fly secrets set ENV=production
fly secrets set ALLOWED_HOSTS=alkime-memos.fly.dev,memos.alki.me
fly secrets set CSP_MODE=strict

# Optional (defaults are production-ready)
fly secrets set HSTS_MAX_AGE=31536000
fly secrets set TRUSTED_PROXIES=10.0.0.0/8
fly secrets set LOG_LEVEL=info
```

## Verification After Deployment

1. **Check server logs:**
   ```bash
   fly logs
   ```
   Verify:
   - No warnings about trusted proxies
   - JSON-formatted logs
   - "env":"production"

2. **Test security headers:**
   ```bash
   curl -I https://alkime-memos.fly.dev/
   ```
   Verify presence of:
   - Strict-Transport-Security
   - X-Frame-Options
   - X-Content-Type-Options
   - Content-Security-Policy

3. **Test health endpoint:**
   ```bash
   curl https://alkime-memos.fly.dev/health
   ```
   Expected: `{"status":"healthy","service":"memos"}`

## Security Header Testing Tools

- [securityheaders.com](https://securityheaders.com/)
- [Mozilla Observatory](https://observatory.mozilla.org/)

Test your deployment:
```
https://securityheaders.com/?q=alkime-memos.fly.dev
```
