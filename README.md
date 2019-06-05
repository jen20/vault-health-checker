## `vault-health-checker`

`vault-health-checker` is a utility which can be run on the same instance as a [Vault][vault] server, and will listen
for TCP  connections from an AWS [Network Load Balancer][nlb] health check only when the Vault server is in an active
or (optionally) standby state, as determined by the Vault health check endpoint.

### Rationale

[HashiCorp Vault][vault] exposes a health check located at the [`/v1/sys/health`] endpoint, which can be used for
determining whether a particular instance is ready for service - in an active or standby state - or whether the instance
is sealed or uninitialized.

HTTP `HEAD` requests to the health check endpoint return a status code representing the health of the target instance
- by default, `200` for an active node, `429` for a standby node, and `500`-range codes for sealed or uninitialized
Vaults.

Unfortunately, the AWS [NLB][nlb] does not support HTTP health checks, instead supporting only TCP checks. While TCP
checks can be pointed at a Vault server, they cannot determine the actual health of the instance, and fill the logs of
the Vault server with spam related to unencrypted requests.

Ideally, the NLB will eventually support HTTP health checks and this project will become obsolete.

### Configuration

`vault-health-checker` is configured using environment variables:

- `VAULT_HEALTH_CHECK_LOG_LEVEL` - The level at which logs should be output. Defaults to `INFO`, which prints a message
  when the health status changes. The other useful value is `DEBUG`, which prints a message for each health check made.

- `VAULT_HEALTH_CHECK_SERVER_ADDR` - The URL of the Vault server to check. This is parsed as a [go-sockaddr][sockaddr]
  template, and should include protocol. The default is `https://{{ GetPrivateIP }}:8200`.

- `VAULT_HEALTH_CHECK_TCP_ADDR` - The address on which to listen for TCP connections in the event that the Vault server
  is healthy. This is parsed as a [go-sockaddr][sockaddr] template. The default is `{{ GetPrivateIP }}:8210`.

- `VAULT_HEALTH_CHECK_INTERVAL` - The interval at which to poll the Vault health check endpoint. This is parsed as a
  [Duration][duration]. The default is `1s` (1 second).

- `VAULT_HEALTH_CHECK_STANDBY_UNHEALTHY` - Whether to consider Vault servers in the `standby` state as unhealthy. If any
  value is assigned to this variable, `standby` nodes are considered unhealthy - by default they are considered healthy.

### Building

During development, `vault-health-checker` can be built using `go build`.

Releases are made using `goreleaser`. Release builds can be invoked using `release.sh` in the root directory of this
repository. `goreleaser` must be available on `PATH`.

### Contributing

Feedback, issues and pull requests are welcome!

[vault]: https://github.com/hashicorp/vault
[nlb]: https://docs.aws.amazon.com/elasticloadbalancing/latest/network/introduction.html
[sockaddr]: https://github.com/hashicorp/go-sockaddr
[duration]: https://golang.org/pkg/time/#ParseDuration
