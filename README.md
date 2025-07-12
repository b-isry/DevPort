# DevPort


**Teleport your `node_modules`. Instant, shareable, and cloud-native.**

[![Release](https://github.com/b-isry/DevPort/actions/workflows/release.yml/badge.svg)](https://github.com/b-isry/DevPort/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/b-isry/DevPort)](https://github.com/b-isry/DevPort/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)


---

## The Problem

Node.js developers suffer from gigantic `node_modules` folders. We waste countless hours on:

- **Slow Installs:** Waiting for `npm install` to download and resolve millions of files.
- **Environment Drift:** "Works on my machine" bugs caused by subtle differences in dependencies.
- **Painful Onboarding:** New team members spend hours setting up a project, only for it to fail.
- **CI/CD Bottlenecks:** CI pipelines reinstall the same dependencies over and over, wasting time and money.

## The Solution: DevPort

**DevPort** is a next-generation dependency engine that virtualizes `node_modules`. It treats your dependencies as a single, versioned artifact tied directly to your `git` commits and stored in a shared cloud cache (like S3).

Instead of running `npm install`, you just run `devport pull`.

DevPort fetches the exact `node_modules` state for your current commit from the cache and reconstructs it on your machine in seconds.


## Features

- **⚡ Instant Dependency Restoration:** Go from `git clone` to a ready-to-run project in seconds, not minutes.
- **☁️ Cloud-Native Caching:** Uses any S3-compatible object store (AWS S3, MinIO, R2) as a shared cache for your team.
- **⚙️ Git-Aware:** The dependency cache is intelligently versioned by your Git commit hashes. Switch branches and `devport pull` to get the right dependencies instantly.
- **✨ Content-Addressable Storage:** Files are deduplicated across all projects and versions, saving significant storage space.
- **🖥️ Cross-Platform:** Aims to seamlessly sync dependencies between macOS, Linux, and Windows environments. (Future Goal)

## Getting Started

### 1. Installation

Download the latest pre-compiled binary for your operating system from the [**GitHub Releases**](https://github.com/b-isry/DevPort/releases/latest) page.

Unzip the archive and place the `devport` binary in a directory that is in your system's `PATH` (e.g., `/usr/local/bin` on macOS/Linux).

### 2. Configuration

DevPort uses two configuration files in your project root.

**A. `.devport.yaml` (Safe to commit)**

This file contains non-secret configuration.

```yaml
# The root directory of the dependencies you want to cache.
root_directory: "node_modules"

# S3-Compatible Cloud Storage Configuration
s3:
"  endpoint: "http://localhost:9000" # For local MinIO, or e.g., "s3.us-west-2.amazonaws.com"
  bucket: "devport-cache"
  region: "us-east-1"
  use_ssl: false # Set to true for most cloud providers
```

**B. `.devport.secret.yaml` (DO NOT commit - add to `.gitignore`)**

This file contains your private credentials.

```yaml
# devport.secret.yaml
s3:
  access_key_id: "YOUR_S3_ACCESS_KEY"
  secret_access_key: "YOUR_S3_SECRET_KEY"
```

**2. The `devport push` Block:**


#### `devport push`

Run this command after you've run `npm install` and committed your changes to `package-lock.json`.

It scans your `node_modules`, uploads any new files to the shared S3 cache, and saves a "manifest" for your current git commit.

```bash
# After running `npm install` and `git commit`...
devport push
```

**3. The `devport pull` Block:**


#### `devport pull`

Run this command after cloning a repo, switching branches, or pulling new changes.

It checks your current git commit, finds the correct manifest in the S3 cache, and reconstructs your `node_modules` folder instantly.

```bash
# After running `git pull` or `git checkout`...
devport pull
```

**4. The `Roadmap` and `Contributing` Lists:**

These sections should use Markdown list syntax (`-` or `*`) to render as bullet points.


## Roadmap

DevPort is currently an MVP. Here's what's next:

- [ ] **Native Module Support:** Robustly handle platform-specific compiled dependencies.
- [ ] **Performance:** Implement concurrent hashing and S3 operations for even faster syncs.
- [ ] **Better Error Handling:** Provide more user-friendly and actionable error messages.

## Contributing

Contributions are welcome! Whether it's bug reports, feature requests, or pull requests, all are appreciated.

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'feat: Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.
