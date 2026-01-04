# Manual Setup for Next Branch

One-time configuration steps to complete in web dashboards:

- [ ] Add `next.yapi.run` custom domain in Vercel dashboard (Settings > Domains)
- [ ] Configure branch alias: `next` branch deploys to `next.yapi.run` (Vercel Settings > Git)
- [ ] Add DNS CNAME record: `next` → `cname.vercel-dns.com`
- [ ] Add branch protection rules for `next` in GitHub repository settings (optional)
- [ ] Create the `next` branch from `main` and push to origin
