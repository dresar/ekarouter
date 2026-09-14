# Credential Rotation Engine

The EkaRouter rotation engine automates failover and intelligent load distribution across pools of third-party credentials.

## Supported Strategies

1. **Priority (`priority`):**
   Selects the credential with the lowest priority number (1 > 2 > 10). If multiple keys share the same priority, the key with fewer requests is selected.

2. **Round Robin (`round_robin`):**
   Evenly distributes incoming requests sequentially across all active, healthy credentials using an atomic counter.

3. **Least Used (`least_used`):**
   Directs the request to the credential with the lowest total request count.

4. **Lowest Error Rate (`lowest_error_rate`):**
   Calculates `error_count / request_count` and chooses the most reliable key.

5. **Health-Based (`health_based`):**
   Scores credentials based on recent latency, validation history, and error logs, selecting top-tier keys first.

6. **Random (`random`):**
   Randomly selects an eligible credential for uniform stateless distribution.

## Concurrency & Race Condition Defense
- Credential selection executes under mutex locks (`sync.Mutex`) in memory.
- Inactive (`status != "active"`), expired (`expires_at < now`), and cooling-down (`cooldown_until > now`) keys are automatically excluded prior to strategy evaluation.
