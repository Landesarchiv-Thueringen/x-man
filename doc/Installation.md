# Installation

## Requirements

x-man runs on any Linux system that provides the following dependencies:

- Docker (configured and with running daemon)
- Docker Compose
- OpenSSL (for testing setup)
- NodeJS and NPM (for frontend development)

Testing on Windows is possible with a Linux virtual machine or using the [Windows Subsystem for Linux (WSL)](https://learn.microsoft.com/en-us/windows/wsl/install) version 2 (WSL 2).

## Getting Started

For a minimal testing setup run the following commands:

```sh
# Create TLS certificates for testing
./scripts/generate-test-certificates.sh
# Create initial configuration
cp .env.example .env
# Build and run for development / testing
docker compose up --build
```

In case you are behind a http proxy, you will need to provide its address with
`HTTP_PROXY` in `.env` before running the last command.

The application will be exposed on http://localhost:8080.

Login:

- fry / fry (user)
- hermes / hermes (administrator)

### Tear Down

```sh
docker compose down --volumes
```

## Test Setup

When following the steps of [Getting Started](#getting-started), some outside
services will be replaced by dummy implementations that are not intended for
production use but allow to create a testing environment without much
configuration.

### LDAP

The development configuration comes with an [OpenLDAP Docker Image for testing](https://github.com/rroemhild/docker-test-openldap).

You can login as any member of the group `ship_crew` and `admin_staff` while the
latter has administration privileges.

### E-Mail

Notification e-mails are sent to a [Mailhog](https://github.com/mailhog/MailHog) instance.

A web UI is available on http://localhost:8025.

### Certificates

The script `generate-test-certificates.sh` creates dummy certificates for the services above.

A root certificate is created and installed in the application container, so it
can verify the certificates and establish secure connections.

## Production Use

### Build and Run

```sh
# Build and run for production
docker compose -f compose.yml -f compose.prod.yml up --build -d
```

Re-run the command after changing configuration.

### Configuration

Copy `.env.example` to `.env` and adjust values as described in the file.

### Custom TLS Certificates

In case any servers you provide in `.env` cannot present certificates that are
signed by a commonly accepted CA, you need to provide any involved root- and
intermediate certificates to x-man.

Copy certificates in PEM format with the file ending `.crt` to
`server/data/ca-certificates` and rerun `docker compose build`. This allows you
to remove the `TLS_INSECURE_SKIP_VERIFY` option for URLs that you provided
certificates for.

Remember to remove the generated `Test-RootCA.crt` for production use.

## Logging

X-man and its services use Docker's [logging mechanism](https://docs.docker.com/config/containers/logging/). View logs via `docker` or `docker compose`, for example

```sh
docker compose logs -f server
```

> [!CAUTION]
> By default, the size of log files is not limited and log files can quickly become very large!

To enable Docker's recommended logging mechanism with a default of 5 rotating files and a maximum of 20 MB, add the following to `/etc/docker/daemon.json`:

```json
{
  "log-driver": "local"
}
```

See https://docs.docker.com/config/containers/logging/configure/ for further information.

## Troubleshooting

### "Verbindung zum Server unterbrochen"

This message indicates that the event stream for client updates from the server
could not be established.
A possible reason is the use of a reverse proxy that does not pass these events.

**Solution (nginx):**

You might need to add some or all of the below options for events to get passed.

For additional information see: https://stackoverflow.com/questions/13672743/eventsource-server-sent-events-through-nginx

```nginx
location / {
    proxy_pass http://url-to-service/;

    # additional options to support server-sent events
    proxy_http_version 1.1;  
    proxy_set_header Connection '';
    chunked_transfer_encoding off;
    proxy_buffering off;
    proxy_cache off;
}
```
