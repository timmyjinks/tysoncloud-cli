# tysoncloud-cli

A personal CLI tool for managing tysoncloud infrastructure and homelab resources. Built for my own use, not intended as a general-purpose tool.

## Overview

`tysoncloud-cli` manages the lifecycle of tysoncloud infrastructure: pulling remote state from Supabase, deploying services to Kubernetes, managing local server records, and migrating legacy deployment data to the current schema.

## Architecture

tysoncloud-cli sits between two data sources and one deployment target:

- **Supabase** is the source of truth for remote infra state - projects, services, environment variables, and volumes all live here.
- **SQLite** is a local database for homelab server records that are managed directly by the CLI and have no Supabase equivalent.
- **Kubernetes** is the deployment target. The CLI talks directly to the cluster via `client-go` using the kubeconfig at `$KUBECONFIG`.

```
                        ┌─────────────────┐
                        │    Supabase     │
                        │  (remote state) │
                        └────────┬────────┘
                                 │ pull / migrate
                                 ▼
┌──────────────┐        ┌─────────────────┐        ┌─────────────────┐
│    SQLite    │◄──────►│ tysoncloud-cli  │───────►│   Kubernetes    │
│ (local only) │        └────────┬────────┘        │    Cluster      │
└──────────────┘                 │                  └─────────────────┘
                                 │ read/write
                                 ▼
                        ┌─────────────────┐
                        │   diff.txt      │
                        │  (local state   │
                        │   snapshot)     │
                        └─────────────────┘
```

### Diff/State File Mechanism

The `pull` command uses a file-based diffing system to determine what has changed in remote infra since the last run. On each pull:

1. The full current state is fetched from Supabase (projects, services, environments, volumes).
2. That state is serialized into a flat text representation where each line is a resource entry.
3. The serialized state is compared against `~/.local/share/tysoncloud/diff.txt`, which holds the last known pulled state.
4. Lines present in the old file but not the new are treated as deletions. Lines present in the new file but not the old are treated as additions.
5. Kubernetes resources are created or destroyed accordingly, then `diff.txt` is updated to reflect the new state.

If any resources fail to create or destroy, they are excluded from the updated `diff.txt` so the next pull will retry them.

Using `--force` skips the diff entirely and redeploys all resources from the current Supabase state.

### Kubernetes Resources

Each service deployed by the CLI results in the following Kubernetes resources being created or updated via server-side apply:

| Resource | Description |
|---|---|
| `Namespace` | Created per project, labelled `managed-by: tysoncloud` |
| `NetworkPolicy` | Restricts ingress/egress per namespace - allows intra-namespace traffic, gateway ingress from `tc-system`, and egress on DNS (53), HTTP (80), and HTTPS (443) |
| `StatefulSet` | Runs the service container with CPU/memory limits and optional volume mounts |
| `Secret` | Holds environment variables for the service, injected via `envFrom` |
| `PersistentVolumeClaim` | Provisioned via the StatefulSet volume claim template if a volume is configured |
| `Service` | ClusterIP service mapping port 80 to the container port |
| `HorizontalPodAutoscaler` | Scales the StatefulSet between 1-10 replicas based on CPU utilization (target 50%) |
| `HTTPRoute` | Gateway API route pointing to `tysoncloud-gateway` in `tc-system`, routing the public domain to the service |

---

## Requirements

- Go 1.25+
- SQLite (via `go-sqlite3`, requires CGO)
- A running Kubernetes cluster with a valid kubeconfig
- A Supabase project

## Environment Variables

| Variable | Description |
|---|---|
| `SUPABASE_URL` | Your Supabase project URL |
| `SUPABASE_API_KEY` | Your Supabase API key |
| `KUBECONFIG` | Path to your kubeconfig file |

## Build

```bash
go build -o tysoncloud ./cmd/cli
```

## Commands

### `pull`

Fetches the current infra state from Supabase and reconciles it against the last known local state (stored at `~/.local/share/tysoncloud/diff.txt`). Creates and destroys Kubernetes resources (namespaces, stateful sets, services, HPAs, HTTP routes, secrets, network policies) to match remote state.

```bash
tysoncloud pull
```

Use `--force` / `-f` to redeploy all resources regardless of diff:

```bash
tysoncloud pull --force
```

---

### `migrate`

Migrates legacy deployment records from the old `deployments` table into the current `projects` + `services` schema in Supabase. Only processes entries with `type = "docker"` and skips anything already present in the new schema.

```bash
tysoncloud migrate
```

---

### `ping`

Pings a server by name (looked up from the local SQLite database) to check reachability. Supports shell autocompletion for server names.

```bash
tysoncloud ping <name>
```

---

### `get`

Fetches and displays resources.

```bash
# List all servers
tysoncloud get servers

# Get a specific server
tysoncloud get servers <name>

# List all projects (from Supabase)
tysoncloud get projects
```

---

### `add`

Adds a new server record to the local SQLite database. Address must be a valid IPv4 address.

```bash
tysoncloud add server <name> <address> [--description <desc>]
```

---

### `update`

Updates an existing server record in the local SQLite database. All flags are optional, only provided fields are updated.

```bash
tysoncloud update server <name> [--name <new-name>] [--description <desc>] [--addr <address>]
```

---

### `delete`

Deletes resources.

```bash
# Delete a server from the local database
tysoncloud delete server <name>

# Delete a deployment namespace from Kubernetes
tysoncloud delete deployment <namespace> <name>
```

---

## Storage

- **Local server records** - SQLite database at `/var/lib/tysoncloud/servers.db`
- **Remote infra state** - Supabase (projects, services, environments, volumes, deployments)
- **Diff state file** - `~/.local/share/tysoncloud/diff.txt` (tracks last known pulled state)

## Database Schema

### SQLite (local - `/var/lib/tysoncloud/servers.db`)

**`servers`** - Homelab server records managed via the CLI.

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `name` | Unique human-readable name |
| `description` | Optional description |
| `addr` | Unique IPv4 address |

### Supabase (remote)

**`projects`** - Top-level grouping for services, mapped to a Kubernetes namespace.

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `user_id` | Owning user |
| `namespace` | Kubernetes namespace name |
| `name` | Project name |
| `created_at` | Creation timestamp |

**`services`** - Individual services belonging to a project.

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `project_id` | Reference to parent project |
| `name` | Service name |
| `resource_name` | Kubernetes resource name |
| `status` | Current service status |
| `url` | Internal URL |
| `public_domain` | Public-facing hostname (used for HTTPRoute) |
| `private_domain` | Internal domain |
| `port` | Container port |
| `image` | Docker image |
| `created_at` | Creation timestamp |

**`environments`** - Key/value environment variables scoped to a service. Deployed as Kubernetes secrets.

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `service_id` | Reference to parent service |
| `key` | Environment variable name |
| `val` | Environment variable value |
| `secret_id` | Optional reference to an external secret |
| `created_at` | Creation timestamp |

**`volumes`** - Persistent volume configuration for a service.

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `service_id` | Reference to parent service |
| `name` | Volume name |
| `storage_gb` | Requested storage in gigabytes |
| `mount_path` | Mount path inside the container |
| `created_at` | Creation timestamp |

**`deployments`** *(legacy)* - Old schema used prior to `projects`/`services`. Only referenced by `migrate`.

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `container_id` | Associated container ID |
| `user_id` | Owning user |
| `name` | Deployment name |
| `url` | Service URL |
| `source` | Docker image source |
| `status` | Deployment status |
| `type` | Deployment type - only `"docker"` entries are migrated |
| `volume` | Volume reference |
| `created_at` | Creation timestamp |

---

## Notes

- Shell autocompletion is supported for commands that accept server or resource names.
- The `pull` command runs creates and destroys concurrently where possible.
