# Development

Use the setup described in [Getting Started](./installation.md#getting-started) as starting point for development.

## Frontend Development Server

To start an auto-refreshing server for frontend development, run

```sh
cd gui
# Install dependencies
pnpm install
# Start the frontend development server
pnpm start
```

## Debugging the Database

The development configuration starts an instance of [mongo-express](https://github.com/mongo-express/mongo-express) and connects it to the application database.

Its web UI is available on [localhost:8081](http://localhost:8081).

## Releasing a New Version

-   Choose a version tag based on semantic versioning.
    In most cases, this means incrementing the minor version when there are new features and otherwise, incrementing the patch version.
-   Update `CHANGELOG.md` with the chosen version tag and any changes.
-   Update the version env in `compose.yml`.
-   Push any changes to `main`.
-   Draft a [new release](https://github.com/Landesarchiv-Thueringen/x-man/releases/new) on GitHub.

## Generating Documentation

We generate documentation with [MkDocs](https://www.mkdocs.org/) and upload the generated site to [GitHub Pages](https://pages.github.com/).

```sh
# Install dependencies (Arch Linux)
sudo pacman -S mkdocs python-dateutil
# Generate and serve docs locally
mkdocs serve
# Generate and upload to GitHub Pages
mkdocs gh-deploy
```

## Server Package Layout

-   `main` (cmd/server.go)  
    Entry point. Not imported by any package.
-   `archive`, `report`, `routines`  
    High-level packages. Only imported by the `main` package.
-   `xdomea`  
    Main application logic. Imported by high-level packages and `main`.
-   `auth`, `errors`, `mail`, `tasks`, `verification`  
    Low-level packages. Imported by higher packages. Depend only on `db` or each other.
-   `db`  
    Database and types. Imported by all other packages. No dependencies to internal packages.

## Error Handling

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
