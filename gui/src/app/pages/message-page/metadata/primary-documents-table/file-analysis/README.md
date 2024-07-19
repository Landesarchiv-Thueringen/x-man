# File Analysis

This collection of Angular components is part of the project [BorgFormat](https://github.com/Landesarchiv-Thueringen/borg).

## Embedding

This feature is designed to be embedded in third-party Angular projects.
To do so,

- Deploy the BorgFormat server on your infrastructure
  - You can omit the `gui` service if not needed
- Copy/paste this directory into your Angular project
- Include `app-file-analysis-table` in your template
- Call the server's endpoint `/analyze-file` with your files and provide the results via the inputs `results` and `getResult` to `app-file-analysis-table`
