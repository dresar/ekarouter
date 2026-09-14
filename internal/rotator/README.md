# Rotator Package

## Purpose
The `rotator` package implements concurrency-safe credential selection and automatic failover across multiple strategies (Priority, Round Robin, Random, Least Used, Lowest Error Rate, and Health-Based).

## Responsibilities
- Concurrency-safe selection preventing multi-request race conditions.
- Automatic filtering of inactive, expired, and cooldown-locked credentials.
- Multi-strategy sorting and selection.

## Public Interfaces
- `NewRotator() *Rotator`
- `(*Rotator) Select(candidates []*vault.Credential, strat Strategy) (*vault.Credential, error)`

## Security Considerations
- Prevents reusing exhausted or compromised keys under active cooldown.
- Rotator operations are executed in memory under mutex protection.
