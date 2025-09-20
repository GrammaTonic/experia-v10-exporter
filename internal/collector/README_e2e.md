Local end-to-end test (Experia V10)
=================================

This package includes an optional end-to-end test that will talk to a real
Experia V10 modem. The test is disabled by default and must be explicitly
enabled via environment variables.

How to run
----------

Set the following environment variables and run `go test` for the package:

EXP E R I A _ E 2 E = 1    # enable the e2e test
EXPERIA_IP=192.168.2.254   # IP address of your modem
EXPERIA_USER=admin         # username
EXPERIA_PASS=secret        # password

Example:

```
EXPERIA_E2E=1 EXPERIA_IP=192.168.2.254 EXPERIA_USER=admin EXPERIA_PASS=secret \
  go test ./internal/collector -run TestE2E -v
```

Notes
-----
- The test will be skipped unless `EXPERIA_E2E` is set to `1` and the other
  variables are present.
- The test avoids printing secrets to stdout/stderr.
- Use this test only on a local machine with physical access to your modem.
