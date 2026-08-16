# Delete your Grom account

This page explains how to permanently delete your **Grom** account and the data stored for it on a Grom server. Use it if you no longer want an account on an instance you signed up for (including after uninstalling the Android app).

Grom is self-hosted: the Android client and web UI connect to **whatever server URL you configured**. Your account and workouts live on that instance. Deletion runs on **that** server — there is no central Grom cloud that holds every user’s data.

## Delete via the web UI (no Android reinstall)

1. Open your instance’s base URL in a browser (the same host you use in the Android app’s server field, for example `https://grom.example.com/`).
2. Sign in with your nickname (or email) and password.
3. Open **Profile**.
4. Open the overflow menu (⋮) and choose **Delete account**.
5. Read the warning, enter your password, and confirm **Delete**.
6. The client shows **Goodbye** and signs you out. The account is gone on that server.

## Delete in the Android app

The steps match the web UI: **Profile** → overflow menu → **Delete account** → enter password → confirm. You must be signed in to the instance where the account exists.

## What is deleted

On successful confirmation, the server permanently removes data associated with your account on that instance, including:

- Login credentials, profile, and avatar
- Your workouts (metadata, GPS tracks, media, map previews, speed and heart-rate charts)
- Equipment
- Follow relationships involving you
- Your likes and comments on other users’ workouts on that instance
- Personal access tokens (PAT)
- Password-reset tokens
- Local federation data for your actor (keys, followers/outbox state stored for your nickname)

The same nickname and email may be registered again on that instance immediately.

## When deletion happens

Deletion is **immediate** after a successful password confirmation. There is no delayed purge queue in Grom.

## Federation

If the instance has ActivityPub federation enabled, Grom sends a `Delete` of your actor to known remote inboxes (**best-effort**) before wiping local data. Remote servers that never receive the activity may keep stale copies of your federated content. Receiving a remote actor `Delete` on this server clears that athlete’s federated inbox entries for local users.

## Requirements and limits

- You must be signed in with a normal session (JWT). Personal access tokens **cannot** delete an account.
- Your current password is required.
- There is no alternate request channel on this documentation site: open the web UI of **your** instance and use **Delete account** as above.
