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
