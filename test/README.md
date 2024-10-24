# End-to-end Tests With Playwright

## Preparation

Build and run the application as described in [Getting Started](https://landesarchiv-thueringen.github.io/x-man/installation/#getting-started).

In this directory, run

```sh
npm install
npx playwright install
```

### Playwright Dependencies

Ubuntu:

```sh
sudo npx playwright install-deps
```

Arch Linux:

```sh
sudo pacman -S nss nspr atkmm cups libdrm libxcomposite libxdamage libxrandr mesa pango alsa-lib libxcursor gtk3
```

## Running Tests

```sh
# Run all tests in headless mode
npx playwright test
# Open Playwright UI
npx playwright test --ui
```

## Writing Tests

```sh
# Open Playwright's interactive code generator
npx playwright codegen http://localhost:8080
```

Visit https://playwright.dev/docs/intro for more information on Playwright.
