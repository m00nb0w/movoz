# Product Spec: Lonely Moley

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13
> **Project**: lonely-moley
> **Location**: `infra/lonely-moley/`
> **Stack**: Kubernetes, Terraform, Ansible, Linux

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
| B1 | Self-hosted CI/CD runner (Gitea Actions, Drone, or Jenkins) | Must have |
| B2 | Build and push container images | Must have |
| B3 | Automated deploy to K8s on push | Should have |
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
| N3 | VPN/tunnel for remote access (Tailscale or WireGuard) | Should have |
| N4 | Firewall and network segmentation | Should have |

### Platform Services

| # | Requirement | Priority |
|---|-------------|----------|
| P1 | Git server (Gitea) as mirror or primary | Nice to have |
| P2 | Container registry (Harbor or registry) | Nice to have |
| P3 | Secret management (Vault or Sealed Secrets) | Should have |

## Hardware

<!-- Fill in your homelab hardware specs -->

| Component | Spec | Notes |
|-----------|------|-------|
| Server(s) | | |
| Storage | | |
| Network | | |

## Migration Plan

The existing `infra/` directory has Terraform for AWS (VPC + RDS) and Docker Compose + Nginx. Over time, Lonely Moley can absorb or replace these:

1. **Phase 1**: Set up K8s cluster, deploy backend services locally
2. **Phase 2**: Self-hosted CI/CD, stop relying on GitHub Actions
3. **Phase 3**: Migrate database from AWS RDS to self-hosted PostgreSQL
4. **Phase 4**: Full self-hosted stack — compute, storage, CI/CD, monitoring

## Open Questions

- [ ] What hardware to start with? (old laptop, Raspberry Pi cluster, mini PC?)
- [ ] k3s vs kubeadm vs Talos Linux?
- [ ] Which CI/CD system? (Gitea Actions, Drone, Tekton?)
- [ ] Keep AWS as disaster recovery / failover, or go fully self-hosted?
- [ ] How to handle dynamic DNS / public access for the homelab?
