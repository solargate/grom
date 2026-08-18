# Approved Grom servers

The Android app can connect to **any** Grom instance: type the server URL on sign-in, registration, or forgot-password. You can also pick from a list of **approved public servers** shipped in the app, plus servers you have already used.

This page is for:

- People using the Android app who want to choose a public instance
- Operators who want their public instance listed in that picker

Listing an instance is **not** an endorsement of that operator’s privacy practices or uptime. Grom has no central cloud; you choose which server holds your account. See the [Privacy Policy](../privacy.md).

## In the Android app

On **Sign in**, **Register**, and **Forgot password**, the server URL field has a dropdown button.

- **Approved servers** come from the catalog compiled into that app version: URL (bold), name, and a short description.
- **Recent servers** are URLs you typed yourself and then **successfully signed in or registered**. They appear below the approved list. Forgot-password does not add a server to this list. There is no delete control; entries rotate out after many other successful connections.
- You can still type any host or URL (scheme optional). HTTP is allowed for local / LAN instances. See [User overview](overview.md) for how the client probes HTTPS then HTTP.

The catalog is **baked into the APK**. Adding a server in git does not update already-installed apps; users see new entries after they install an Android release that includes the change.

## Add your instance (pull request)

Anyone can open a pull request against [`server-catalog.yaml`](https://github.com/solargate/grom/blob/master/server-catalog.yaml) in the repository root.

1. Fork [solargate/grom](https://github.com/solargate/grom) and create a branch.
2. Add **one** entry under `servers:` (do not remove or rewrite other instances).
3. Open a pull request to `master`. The maintainer reviews it.

Example:

```yaml
servers:
  - url: https://grom.example.org
    name: Example Grom
    description: Public instance for testing federation.
    email: admin@example.org
```

### Fields

| Field | Shown in the app | Rules |
|-------|------------------|--------|
| `url` | Yes (bold) | Must start with `https://`. **No port** (not even `:443`). A path is allowed (`https://host/grom`). No userinfo, query, or fragment. Hostname must be a DNS name (not an IP). |
| `name` | Yes | Short display name, at most 80 characters. One language, as you write it. |
| `description` | Yes (small text) | One or two sentences, at most 280 characters. One language, as you write it. |
| `email` | No | Operator contact for review. Must contain `@`. |

CI rejects invalid YAML (missing `https://`, ports, duplicates after normalizing trailing slashes, and so on). Maintainers may also reject listings that are not a public Grom instance, have a broken TLS certificate, or look misleading.

After merge, the instance is included in the **next Android release**. Run `make catalog` if you also need to refresh the generated Flutter file in the same change; CI fails if that file is out of date.

## See also

- [User overview — sign-in](overview.md#sign-in-and-password-reset)
- [Install and run](../admin/install.md) (operators connecting the Android app to their own server)
