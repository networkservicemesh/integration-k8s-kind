
# Test use-cases

Prepare DNS and Spire:
```bash
./performance_testing/scripts/nsm_setup_dns.sh &&
./performance_testing/scripts/nsm_setup_spire.sh
```

Test with vl3 DNS:
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
