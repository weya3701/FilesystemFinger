# FilesystemFinger

`fsfinger` creates deterministic SHA-256 fingerprints for files and directory
trees. A directory fingerprint changes when a relative path, object type,
symlink target, empty directory, or file content changes. Moving the scanned
root itself does not change the fingerprint.

Metadata such as permissions and modification time is included in the JSON
manifest for inspection, but never contributes to the root hash.

## Build and use

```sh
go build -o fsfinger ./cmd/fsfinger

./fsfinger scan /path/to/directory
./fsfinger scan --hash-only /path/to/directory
./fsfinger scan /path/to/directory > manifest.json
./fsfinger verify --manifest manifest.json /path/to/directory
```

Write the manifest outside the scanned directory, or ignore its path, so the
output does not become part of the next scan.

## Ignore rules

Built-in rules ignore `.git/`, `.DS_Store`, `Thumbs.db`, `*.swp`, and `*~`.
Add repeatable patterns with `--ignore`, load rules from a file with
`--ignore-file`, or disable defaults with `--no-default-ignore`.

Patterns use `/` separators on every platform. `*` matches within a path
segment, `**` spans directories, `?` matches one character, a trailing `/`
matches directories only, and a leading `!` negates an earlier rule.

## Stable format

Names and symlink targets are normalized to Unicode NFC and paths use `/`.
Directory children are sorted by normalized UTF-8 bytes. Canonical records are
length-prefixed, use big-endian lengths, and are domain-separated with the
`filesystem-fingerprint-v1` format identifier. Files are hashed as their exact
bytes; line endings are not rewritten. Symlinks are not followed.
