
# Performance testing

This folder contains deployment yaml files and scripts
that deploy, run and clear applications for performance testing.

# Parameters

Parameters to be considered are:

1. `qps_list`: requested load of the system
2. `duration`: duration of a single test
3. `connections`: the amount of simultaneous connections from test client to test server
4. `iterations`: how many times to run each test

To inspect results you can install Fortio and run `fortio server`.
In the web ui you will be able to see graphs for different runs and compare them.

Alternatively you can simply open .json files and inspect them for QPS and different latency percentiles.

# Running the tests manually locally

Make sure that you have load ballancer in you cluster.
For Kind and bare metal clusters you can use metallb installation script:
```bash
./performance_testing/scripts/setup_metallb.sh
```

Prepare DNS and Spire:
```bash
./performance_testing/scripts/nsm_setup_dns.sh &&
./performance_testing/scripts/nsm_setup_spire.sh
```

Test interdomain vl3:
```bash
./performance_testing/scripts/run_test_suite.sh \
    vl3 \
    ./performance_testing/results/raw/ \
    3 \
    "http://nginx.my-vl3-network:80" \
    "./performance_testing/use-cases/vl3/deploy.sh" \
    "./performance_testing/use-cases/vl3/clear.sh" \
    "v1.8.0" \
    "./performance_testing/nsm"
```

Test interdomain wireguard:
```bash
./performance_testing/scripts/run_test_suite.sh \
    k2wireguard2k \
    ./performance_testing/results/raw/ \
    3 \
    "http://172.16.1.2:80" \
    "./performance_testing/use-cases/k2wireguard2k/deploy.sh" \
    "./performance_testing/use-cases/k2wireguard2k/clear.sh" \
    "v1.8.0" \
    "./performance_testing/nsm"
```

Clear cluster if needed:
```bash
./performance_testing/scripts/nsm_clear_spire.sh
./performance_testing/scripts/nsm_clear_dns.sh
```
