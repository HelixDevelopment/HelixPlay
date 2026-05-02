# HelixPlay Operations Runbook

## Alerting Playbooks

### HighLatency (p99 > 50ms)

1. Check `helixplay_latency_p99_ms` metric
2. Identify region from label
3. Scale out streamers in affected region
4. If persistent, check network path for congestion

### StreamerHighErrorRate (> 1%)

1. Check `helixplay_stream_errors_total` counter
2. Examine host agent logs for codec errors
3. If GPU errors, restart host agent pod
4. If network errors, verify MTU and firewall rules

### DiscoveryStaleHosts (> 5min since heartbeat)

1. Check `helixplay_discovery_stale_hosts` gauge
2. Verify NATS event bus connectivity
3. If network partition, enable split-brain recovery
4. Manually deregister dead hosts if needed

## Scaling Procedures

### Horizontal Scale-Out

```bash
# Add streamer replicas
kubectl scale deployment helixplay-core --replicas=10 -n helixplay

# Add host agents
kubectl patch daemonset helixplay-host-agent \
  -n helixplay -p '{"spec":{"template":{"spec":{"nodeSelector":{"gpu":"true"}}}}}'
```

### Database Maintenance

```bash
# CockroachDB rolling restart
kubectl rollout restart statefulset cockroachdb -n helixplay

# Redis failover test
kubectl delete pod redis-0 -n helixplay
```

## Disaster Recovery

### Region Failure

1. DNS failover to standby region
2. NATS cross-region replication ensures event continuity
3. CockroachDB multi-region survival ensures data consistency
4. Host agents in failed region will reconnect to new rendezvous

### Data Corruption

1. Stop all writes to affected tables
2. Restore from CockroachDB backup (S3)
3. Verify checksums before bringing service back online

## Security Incident Response

### Compromised Host Agent

1. Revoke mTLS certificate via CA
2. Deregister host from discovery
3. Isolate node at network level
4. Forensic analysis of host filesystem

### DDoS Mitigation

1. Enable rate limiting at edge (Cloudflare/AWS Shield)
2. Scale core replicas horizontally
3. Enable circuit breakers on discovery service
4. Alert on-call engineer if > 10k RPS sustained
