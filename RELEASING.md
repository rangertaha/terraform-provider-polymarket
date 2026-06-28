# Releasing

This provider publishes to the [Terraform Registry](https://registry.terraform.io)
via a tag-triggered GitHub Actions workflow
([`.github/workflows/release.yml`](./.github/workflows/release.yml)) that runs
[GoReleaser](https://goreleaser.com) and GPG-signs the artifacts. The registry
requires every release to be signed, so a GPG key is a one-time prerequisite.

## 1. Generate a GPG signing key

You need `gpg` installed (`gpg --version`). Generate a key dedicated to signing
releases:

```sh
gpg --full-generate-key
```

Answer the prompts:

- **Key type:** `(1) RSA and RSA`
- **Key size:** `4096`
- **Expiry:** `0` (does not expire) — or set one and remember to rotate it
- **Real name / email:** your name and the email tied to your GitHub/registry
  account
- **Passphrase:** choose a strong one and keep it — it becomes a repo secret

Non-interactive alternative:

```sh
cat > key.params <<'EOF'
%echo Generating release signing key
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: Your Name
Name-Email: you@example.com
Expire-Date: 0
%commit
EOF
gpg --batch --pinentry-mode loopback --passphrase 'YOUR_PASSPHRASE' --gen-key key.params
rm key.params
```

Find the key's fingerprint (the long hex string after `sec`):

```sh
gpg --list-secret-keys --keyid-format=long
# sec   rsa4096/ABCD1234EF567890 2026-06-28 [SC]
#       4F2E....FULL_FINGERPRINT....A1B2   <-- this line
```

## 2. Add the public key to the Terraform Registry

Export the **public** key and register it so the registry can verify your
signatures:

```sh
gpg --armor --export <FINGERPRINT>
```

Copy the output (including the `-----BEGIN/END PGP PUBLIC KEY BLOCK-----` lines)
into the registry at
**[registry.terraform.io/settings/gpg-keys](https://registry.terraform.io/settings/gpg-keys)**.

## 3. Add repository secrets

The release workflow reads two secrets. Export the **private** key:

```sh
gpg --armor --export-secret-keys <FINGERPRINT>
```

Add both secrets under **GitHub → repo → Settings → Secrets and variables →
Actions → New repository secret**:

| Secret name       | Value                                                        |
| ----------------- | ----------------------------------------------------------- |
| `GPG_PRIVATE_KEY` | the full ASCII-armored private key block from the export    |
| `PASSPHRASE`      | the passphrase you chose in step 1                          |

Or with the GitHub CLI:

```sh
gpg --armor --export-secret-keys <FINGERPRINT> | gh secret set GPG_PRIVATE_KEY
gh secret set PASSPHRASE   # paste the passphrase when prompted
```

> The default `GITHUB_TOKEN` the workflow uses for creating the release is
> provided automatically — you do not need to add it.

## 4. Claim the provider on the registry

Sign in to the [Terraform Registry](https://registry.terraform.io) with GitHub,
choose **Publish → Provider**, and select this repository. The repository name
must be `terraform-provider-<name>` (it is) and you must have added a GPG key in
step 2.

## 5. Cut a release

1. Update [`CHANGELOG.md`](./CHANGELOG.md): move items from `Unreleased` into a
   new version section with today's date.
2. Commit the changelog.
3. Tag and push — the tag triggers the release workflow:

   ```sh
   git tag -a v0.1.0 -m "v0.1.0"
   git push origin v0.1.0
   ```

GoReleaser builds the cross-platform binaries, signs the checksums with your GPG
key, and creates the GitHub Release. The Terraform Registry picks up the new
release automatically (usually within a few minutes).

### Verifying / re-running

- Watch the run under the repo's **Actions → Release** tab.
- If a release fails (e.g. a missing secret), fix it, then delete and recreate
  the tag:

  ```sh
  git push --delete origin v0.1.0
  git tag -d v0.1.0
  # ...fix, then re-tag and push again
  ```

- To dry-run the build locally without publishing:

  ```sh
  goreleaser release --snapshot --clean
  ```
