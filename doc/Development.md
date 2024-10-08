# Development

When run with development configuration (see [Getting Started](./Installation.md#getting-started)), there are some additional options to help with testing and debugging.

## Frontend Development Server

To run with a auto-refreshing development server for frontend development, run

```sh
# Run a minimal backend configuration. You can also start the complete stack without specifying "server".
docker compose up --build -d server
# Start the frontend development server
cd gui
npm start
```

## Debug the Database

The development configuration starts an instance of [mongo-express](https://github.com/mongo-express/mongo-express) and connects it to the application database.

Its web UI is available on http://localhost:8081.

## Concepts

### Error Handling

**Error and panic.**
In the server, we use a combination of go `error` return values and `panic`. In general, expected problems—such as connection issues or invalid files—should be returned as `error` while unexpected problems due to programming errors can `panic`.
When returning an `error`, take care to provide enough context, so the problem can be clearly identified.

Also consider wether operation can continue after the error. For example, for an error sending a notification email, it might be better not to `panic` to not abort any subsequent steps.

**ProcessingError.**
Either is turned in a `ProcessingError` and displayed in the administration UI.

**Error on misconfiguration.**
In general, it is ok to `panic` on missing or invalid environment variables. However, take care to gracefully handle any connection issues with `error` returns.

**Recovering from a panic.**
`panic`s are recovered from to not crash the application. This happens by Gin when handling HTTP requests and should be taken care of by the programmer when invoking a goroutine.
Take care to not cause further `panic`s when recovering from a previous `panic`, since this might crash the application.

## Using Docker-Compose

There are three compose files serving different purposes:

- `compose.yml` is the base file for all other files. On its own, it sets up the production runtime with existing images, that have to be built before.
- `compose.override.yml` contains adaptions for development and testing. It will be included by docker-compose automatically when present. It contains build instructions, configuration adaptations, and additional services.
- `compose.build-prod.yml` contains build instructions for production builds. It has to be specified when calling docker-compose explicitly. After building images using this file, a production setup can be stared using `compose.yml` alone.

See [Getting Started](./Installation.md#getting-started) and [Build and Run](./Installation.md#build-and-run) for instructions how to use these files.
