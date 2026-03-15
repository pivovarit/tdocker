# ADR-0002: Docker CLI over Docker SDK

**Date:** 2026-03-15

## Context and Problem Statement

tdocker interacts with Docker by calling the  `docker` CLI directly and parsing its JSON output. An alternative is the Docker Engine SDK (`github.com/docker/docker/client`), which provides a native Go client for the Docker API with typed requests and responses.

The question was whether to replace the CLI wrapper (`internal/docker/`) with the SDK.

## Considered Options

* **Docker SDK** - use `github.com/docker/docker/client` to talk directly to the Docker Engine API over the Unix socket.
* **Docker CLI wrapper** - continue shelling out to `docker` and parsing JSON output (current approach).

## Decision Outcome and Drivers

Chosen option: **Docker CLI wrapper**, because:

* **Feature coverage** - several features tdocker relies on have no SDK equivalent: `docker debug` (Docker Desktop plugin), `docker exec -it` with TTY passthrough, `docker logs --grep` (server-side log filtering), and `docker context ls/use` (context switching reads `~/.docker/config.json` and CLI handles negotiation transparently). Adopting the SDK would still require shelling out for these, resulting in two integration layers instead of one.
* **Dependency weight** - the Docker SDK pulls in a large transitive graph (containerd, opencontainers, gRPC). This would significantly increase binary size and the dependency surface for what amounts to a thin wrapper over a handful of commands.
* **API version negotiation** - the CLI auto-negotiates the API version with the daemon. The SDK requires the caller to manage version compatibility, adding complexity for no user-facing benefit.
* **Simplicity** - the current approach is ~1600 LOC in `internal/docker/`, each file handling one concern (containers, logs, events, exec, stats, inspect, contexts). The JSON output from `docker` is stable and well-documented. The code is straightforward to test with stub clients.

The SDK would be the right choice for a project that needs fine-grained programmatic control (building images, managing networks, watching container health transitions). However, tdocker is a read-heavy TUI over `docker ps` with a handful of lifecycle commands.

## People
- @pivovarit
