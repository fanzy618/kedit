# kedit

`kedit` is a command‑line utility for managing Kubernetes `kubectl` configuration files (*kubeconfig*). It lets you list, delete, rename, prune and merge clusters, users and contexts, so that switching between Kubernetes clusters becomes effortless.

## Features

* **List items** — display the names of clusters, users or contexts, or list all at once.
* **Delete items** — remove specific clusters, users or contexts.
* **Rename items** — rename clusters, users or contexts and automatically update all references, including the `current-context`.
* **Prune config** — remove clusters and users that are not referenced by any context.
* **Merge contexts** — import a context (together with its cluster and user) from one kubeconfig file into another.
* **Flexible target** — work on a user‑specified kubeconfig file or default to `$HOME/.kube/config`.

## Getting Started

### Prerequisites

* Go ≥ 1.18 installed.
* The `kedit` source code.

### Building from source

```bash
go build -o kedit .
```

This produces an executable named **kedit** (or **kedit.exe** on Windows) in the current directory.

### Installing the binary (optional)

#### Linux / macOS

```bash
sudo mv kedit /usr/local/bin/
```

(or copy it to any directory included in your `PATH`, e.g. `~/bin`).

#### Windows

Move `kedit.exe` to a directory that is part of your *Path* environment variable.

## Usage

### Global options

```
-k, --kubeconfig <FILE_PATH>   Path to the kubeconfig file to operate on
                               (default: $HOME/.kube/config)
```

### Commands

Below is a quick reference. Run `kedit <command> --help` for the full syntax of each command.

#### list

List the names of clusters, users, contexts, or all of them at once.

```bash
kedit list cluster
kedit list user
kedit list context
kedit list all
```

#### delete

Delete a cluster, user or context.

```bash
kedit delete cluster <cluster-name>
kedit delete user <user-name>
kedit delete context <context-name>
```

#### rename

Rename a cluster, user or context (all references are updated automatically).

```bash
kedit rename context <old-name> <new-name>
```

#### prune

Remove unused clusters and users from the kubeconfig.

```bash
kedit prune
```

#### merge

Merge a context (with its cluster and user) from another kubeconfig.

```bash
kedit merge <context-name> --from /path/to/other/kubeconfig [--name <new-name>]
```

---

*Happy Kubernetes hacking!*
