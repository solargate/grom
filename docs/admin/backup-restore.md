# Backup and restore

A server backup covers both metadata and files such as tracks, photos, avatars, and keys. Keep them from the same point in time.

## Back up {#backup}

1. Stop Grom so that no workouts or imports can change the data while you copy it. Record the Grom version and the paths used by the running service.
2. Copy the configuration file and the entire resolved `storage.location` directory. This includes the filesystem data for both drivers, federation keys, and the VAPID key pair for push notifications.
3. For `bbolt`, also copy the database if `storage.bbolt.path` places it outside `storage.location`. The default database is `grom.db` inside that directory. If you use static TLS, include the certificate and private key referenced by the configuration.
4. Store the copy securely. It contains private workout data and secrets. Start Grom again when the copy is complete.

Keep temporary Strava uploads outside the backup when `storage.temp_dir` is separate from the data directory. After an interruption, users can upload the same archive again.

## Restore {#restore}

1. Stop Grom and keep a separate copy of the current data before replacing it.
2. Restore the configuration, data directory, and any external database or TLS files from the same backup. Use the original Grom version for the initial restore. Preserve the storage driver, configured federation hostname, and file permissions. If the binary directory changes, check where the relative paths now resolve.
3. Start Grom with the restored configuration. Sign in, open your workout list and a workout with a track and photos, and check the track and attachments. If possible, test recovery in a separate environment before relying on the backup.

See [Configuration](configuration.md#storage-drivers) for driver settings and migration.
