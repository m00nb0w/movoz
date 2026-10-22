# Product Spec: Lonely Moley

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13
> **Last updated**: 2026-10-22
> **Project**: lonely-moley
> **Location**: `infra/lonely-moley/`
> **Stack**: Kubernetes, Terraform, Ansible, Linux
> **Technical design (v1)**: [wiki/technical/lonely-moley-design.md](../technical/lonely-moley-design.md)

## Overview

Lonely Moley is a personal homelab — a self-built cloud platform that replaces reliance on public cloud providers. It hosts the infrastructure and services that power the other movoz projects: Kubernetes clusters, CI/CD build systems, databases, and anything else that would otherwise run on AWS/GCP.

## Problem Statement

Public cloud is expensive for personal projects and limits learning. Running your own infrastructure gives full control, deeper understanding of the stack, and zero recurring cloud bills beyond hardware and electricity. It also serves as a playground for learning infrastructure engineering hands-on.

## Goals

- Self-hosted Kubernetes cluster for running movoz services
- CI/CD pipeline for building and deploying projects without GitHub Actions / cloud CI
- Replace or complement existing AWS infrastructure (Terraform/RDS) over time
- Learn infrastructure engineering by owning every layer

## Non-Goals

- Not production-grade HA (acceptable trade-off for personal use)
- Not a hosting provider for others
- Not replacing all cloud services immediately — gradual migration

## Components

### Compute — Kubernetes

| # | Requirement | Priority |
|---|-------------|----------|
| C1 | K8s cluster (k3s or kubeadm) on bare metal / homelab hardware | Must have |
| C2 | Deploy movoz backend services (hustle-turtle, drunken-dolphin) | Must have |
| C3 | Ingress controller for routing traffic to services | Must have |
| C4 | Persistent storage (local-path or NFS) | Must have |
| C5 | Cluster monitoring (Prometheus + Grafana) | Should have |

### CI/CD — Build System

| # | Requirement | Priority |
|---|-------------|----------|
| B1 | Self-hosted CI/CD runner (Gitea Actions, Drone, or Jenkins). Must not run code from public pull requests. | Nice to have |
| B2 | Build and push container images | Must have |
| B3 | Automated deploy on push (Docker Compose in v1, K8s later) | Should have |
| B4 | Build caching for faster pipelines | Nice to have |

### Data — Databases & Storage

| # | Requirement | Priority |
|---|-------------|----------|
| D1 | PostgreSQL instance for hustle-turtle and other services | Must have |
| D2 | Automated database backups | Must have |
| D3 | Object storage (MinIO) for files/assets | Nice to have |

### Networking

| # | Requirement | Priority |
|---|-------------|----------|
| N1 | Reverse proxy with TLS (Traefik or Nginx) | Must have |
| N2 | DNS management for homelab domain | Must have |
| N3 | VPN/tunnel for remote access (Tailscale or WireGuard) | Must have |
| N4 | Firewall and network segmentation | Must have |
| N5 | Public access with no open router ports and the home IP hidden | Must have |

### Platform Services

| # | Requirement | Priority |
|---|-------------|----------|
| P1 | Git server (Gitea) as mirror or primary | Nice to have |
| P2 | Container registry (Harbor or registry) | Nice to have |
| P3 | Secret management (Vault or Sealed Secrets) | Should have |

## Hardware

| Component | Spec | Notes |
|-----------|------|-------|
| Server(s) | Lenovo ThinkCentre (one box) | Ubuntu Server LTS. Single point of failure, accepted. |
| Storage | Internal disk, LUKS encrypted | Offsite encrypted backups to B2 or S3 |
| Network | Home router, box on its own VLAN | No port forwards. Small UPS for power cuts. |

## Migration Plan

The existing `infra/` directory has Terraform for AWS (VPC + RDS) and Docker Compose + Nginx. Over time, Lonely Moley can absorb or replace these:

1. **Phase 1 (v1)**: Harden the ThinkCentre. Run services and PostgreSQL with Docker Compose. Serve public apps through Cloudflare Tunnel. Deploy from GitHub Actions over Tailscale. See the [v1 technical design](../technical/lonely-moley-design.md).
2. **Phase 2**: Move from Docker Compose to k3s.
3. **Phase 3**: Replace Cloudflare Tunnel with a VPS + WireGuard tunnel.
4. **Phase 4**: Self-hosted CI/CD, if it can run safely with a public repo.
5. **Phase 5**: Full self-hosted stack — compute, storage, CI/CD, monitoring.

## Open Questions

- [x] What hardware to start with? → One Lenovo ThinkCentre.
- [x] How to handle dynamic DNS / public access for the homelab? → Cloudflare Tunnel in v1. No open ports, so no dynamic DNS is needed. VPS tunnel later.
- [x] Compose or K8s first? → Docker Compose in v1. k3s in Phase 2.
- [ ] k3s vs kubeadm vs Talos Linux?
- [ ] Which CI/CD system? (Gitea Actions, Drone, Tekton?) The movoz repo is public, so a runner must not run code from public pull requests.
- [ ] Keep AWS as disaster recovery / failover, or go fully self-hosted?
