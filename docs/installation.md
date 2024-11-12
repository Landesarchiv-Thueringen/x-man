# Installation

## Download

Download the [latest release](https://github.com/Landesarchiv-Thueringen/x-man/releases/latest) from GitHub or clone the [repository](https://github.com/Landesarchiv-Thueringen/x-man) if you want the development version.

## Requirements

x-man runs on any Linux system that provides the following dependencies:

-   Either
    -   Docker (configured and with running daemon), or
    -   Podman
-   Docker Compose (for both, Docker and Podman)
-   OpenSSL (for testing setup)
-   NodeJS and PNPM (for frontend development)

Testing on Windows is possible with a Linux virtual machine or using the [Windows Subsystem for Linux](https://learn.microsoft.com/en-us/windows/wsl/install) version 2 (WSL 2).

## Services

x-man is meant to be used in combination with some external services. For
testing purposes, we provide configuration for dummy services (see [Test Setup](#test-setup)). For production use, you should provide and configure these services (see [Configuration](#configuration)).

The following external services should be provided for use in production:

-   LDAP (user authentication)
-   DIMAG (archiving)
-   [Borg](https://github.com/Landesarchiv-Thueringen/borg) (format verification)
-   SMTP (e-mail notifications)

See [Betriebshandbuch (de)](./betriebshandbuch.md) for further details.

### Compatible Borg Versions

| x-man Version | Borg Version |
| ------------- | ------------ |
| 0.9.x         | 1.0.x        |
| >= 1.0.0      | >= 1.1.0     |

## Getting Started

For a minimal testing setup run the following commands:

```sh
# Create TLS certificates for testing
./scripts/generate-test-certificates.sh
# Create the initial configuration
cp .env.example .env
# Activate the development / testing setup
ln -s compose.dev.yml compose.override.yml
# Build and run
docker compose up --build -d
```

In case you are behind a http proxy, you will need to provide its address with
`HTTP_PROXY` in `.env` before running the last command.

The application will be exposed on [localhost:8080](http://localhost:8080).

### Login

| User   | Password | Role      |
| ------ | -------- | --------- |
| fry    | fry      | Archivist |
| hermes | hermes   | Admin     |

### Tear Down

```sh
docker compose down --volumes
```

## Using Docker-Compose

There are compose files for different purposes:

-   `compose.yml` is the base file for all other files. On its own, it sets up the production runtime with existing images or builds images for production.
-   `compose.dev.yml` contains adaptions for development and testing. It contains build- and configuration adaptations and additional services.
-   `compose.override.yml` will be included by docker-compose automatically when present. It is not present in the repository, but you can link `compose.dev.yml` to it or create your own for custom changes.

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

## Configuration

Copy `.env.example` to `.env` and adjust values as described in the file.

```sh
# Create the initial configuration
cp .env.example .env
# Adjust values as needed
$EDITOR .env
```

Re-run `docker compose up -d` after changing configuration.

## Production Use

Create and adjust an `.env` file as described above. The production setup requires you to provide valid configuration for the required [services](#services).

```sh
# Remove overrides for development setup if any
rm -f compose.override.yml
# Build
docker compose build
# Run
docker compose up -d
```

For the last command, you do not need the entire repository, but only these two files:

-   .env
-   compose.yml

## Custom TLS Certificates

In case any servers you provide in `.env` cannot present certificates that are
signed by a commonly accepted CA, you need to provide any involved root- and
intermediate certificates to x-man.

Copy certificates in PEM format with the file ending `.crt` to
`data/ca-certificates` and rerun `docker compose up`. This allows you
to remove the `TLS_INSECURE_SKIP_VERIFY` option for URLs that you provided
certificates for.

Remember to remove the generated `Test-RootCA.crt` for production use.

## Logging

X-man and its services use Docker's [logging mechanism](https://docs.docker.com/config/containers/logging/). View logs via `docker` or `docker compose`, for example

```sh
docker compose logs -f server
```

!!! warning

    By default, the size of log files is not limited and log files can quickly become very large!

To enable Docker's recommended logging mechanism with a default of 5 rotating files and a maximum of 20 MB, add the following to `/etc/docker/daemon.json`:

```json
{
    "log-driver": "local"
}
```

See [Configure logging drivers](https://docs.docker.com/config/containers/logging/configure/) for further information.

## Troubleshooting

### "Verbindung zum Server unterbrochen"

This message indicates that the event stream for client updates from the server
could not be established.
A possible reason is the use of a reverse proxy that does not pass these events.

**Solution (nginx):**

You might need to add some or all of the below options for events to get passed.

For additional information see: [stackoverflow: EventSource / Server-Sent Events through Nginx](https://stackoverflow.com/questions/13672743/eventsource-server-sent-events-through-nginx).

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
