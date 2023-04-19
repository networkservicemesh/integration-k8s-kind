
# Known results

This file contains info about results we already have.

There are several different QPS targets. Each target has its own result expectations.

# NSM v1.8.0, vl3

vl3 tests in v1.8.0 seems to be CPU throttled by github, which affects max latency.

1. Target QPS == 100
    Actual QPS: 100
    Min latency: 0.3-0.35 ms
    Max latency: 100-250 ms
    Avg latency: 4-5 ms
    p50 latency: 1.5-2.5 ms
    p99 atency: 90-150 ms
2. Target QPS == 1000
    Actual QPS: 350-450
    Min latency: 0.25-0.3 ms
    Max latency: 100-300 ms
    Avg latency: 2-2.5 ms
    p50 latency: 0.7-1.3 ms
    p99 atency: 40-80 ms
3. Target QPS == 1000000
    Actual QPS: 350-450
    Min latency: 0.25-0.3 ms
    Max latency: 100-300
    Avg latency: 2-2.5 ms
    p50 latency: 0.7-1.3 ms
    p99 atency: 40-80 ms

# NSM v1.8.0, wireguard

1. Target QPS == 100
    Actual QPS: 100
    Min latency: 0.3 ms
    Max latency: 20-50 ms
    Avg latency: 1-2 ms
    p50 latency: 0.6 ms
    p99 atency: 15-35 ms
2. Target QPS == 1000
    Actual QPS: 1000
    Min latency: 0.2 ms
    Max latency: 30-50 ms
    Avg latency: 0.8 ms
    p50 latency: 0.4-0.5 ms
    p99 atency: 12-15 ms
3. Target QPS == 1000000
    Actual QPS: 1200-1400
    Min latency: 0.2 ms
    Max latency: 40-50 ms
    Avg latency: 0.7 ms
    p50 latency: 0.4 ms
    p99 atency: 12-15 ms

