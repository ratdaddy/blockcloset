# Shared Packages

Reusable Go packages shared across Block Closet services.

## Structure

- `validation`: S3-compatible validators for bucket names and object keys, exposed via `DefaultBucketNameValidator` and `DefaultKeyValidator`.

## Development

### entr

`entr` is used to automatically restart the server and tests when files change.
Install `entr` with:

```bash
brew install entr
```
