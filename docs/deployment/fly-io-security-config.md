# Fly.io Security Configuration

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
