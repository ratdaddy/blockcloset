# Shared Packages

Reusable Go packages shared across Block Closet services.

## Structure

- `storage/bucket`: S3-compatible bucket name validation exposed via `DefaultBucketNameValidator`.

## Development

### entr

`entr` is used to automatically restart the server and tests when files change.
Install `entr` with:

```bash
brew install entr
```
