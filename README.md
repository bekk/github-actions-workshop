# GitHub Actions workshop

## Requirements

* A GitHub account, with a functioning SSH-key or password setup
* Available GitHub action minutes and storage space (included in free tier)

The tasks in the workshop can be done using only the built-in GitHub editor. However, in order to learn about supporting tooling, and simplify some tasks, the workshop will assume you've installed the following tools:

* Your preferred terminal emulator/shell and `git`
* The [GitHub CLI](https://github.com/cli/cli#installation)
* [actionlint](https://github.com/rhysd/actionlint/blob/main/docs/install.md)
* [ShellCheck](https://github.com/koalaman/shellcheck?tab=readme-ov-file#installing) (will be used by actionlint)
* Editor with YAML and GitHub actions plugins (e.g., VS Code with the [YAML](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) and [GitHub Actions](https://marketplace.visualstudio.com/items?itemName=github.vscode-github-actions) extensions)


## Getting started

TODO: Likely `gh repo fork --clone` or using the GitHub UI to fork the repo to a personal account

This repository contains a simple go app. You do not need to know go, nor use any Golang tooling. We will, unless explicitly specified otherwise, only modify files in the special `.github/` directory.

## Our first workflow

1. We'll start with a simple workflow. Create the file `.github/workflows/test.yml` and with the following content:

    ```yml
    # The "display name", shown in the GitHub UI
    name: Build and test

    # Trigger, run on push on any branch
    on:
      push:

    jobs:
      test: # The 'build' job
        name: "Build application"
        runs-on: 'ubuntu-latest'
        steps:
          # Step to print a simple message
          - run: echo "Hello world"
    ```

    `.github/workflows` is a special directory where all workflows should be placed.

2. Before committing and pushing, run `actionlint` from the repository root. It should run with a zero (successful) exit code and no output, since the workflow file is without errors. Try again with `actionlint --verbose` and verify the output to confirm that it found your file. By default, `actionlint` scans all files in `.github/workflows/`

3. Commit and push the workflow to your fork. In the GitHub UI, navigate to the "Actions" tab, and verify that it runs successfully.

> [!NOTE] 
> *How does this work?*
> 
> Let's break down the workflow file.
> 
> * `name:` is only used for display in the GitHub UI
> * `on:` specifies triggers - what causes this workflow to be run
> * `jobs:` specifies each _job_ in the workflow. A job runs on a single virtual machine with a given OS (here: `ubuntu-latest`), and the `steps` share the environment (filesystem, installed tools, environment variables, etc.). Different jobs have separate environments.
> * `steps:` run sequentially, and might run shell scripts or an action (a reusable, pre-made piece of code). Each step can run conditionally. If a step fails, all later steps fail by default (this is overridable).


## Build and test the application

1. Let's use some pre-made actions to checkout our code, and install Golang tooling. Replace the "hello world" step with the following steps:

    ```yml
          # Checkout code
          - uses: actions/checkout@v4

          # Install go 1.21
          - name: Setup go
            uses: actions/setup-go@v4
            with: # Specify input variables to the action
              go-version: '1.21.x'

          # Shell script to print the version
          - run: go version
    ```

2. Again, run `actionlint` before you commit and push. Verify that the correct version is printed.

3. Continue by adding steps to build and test the application:

    ```yml
      - name: Build
        run: go build -v ./...

      - name: Test
        run: go test ./...
    ```

4. Verify that the workflow fails if the build fails (create a syntax error in any file). Separately, verify that the workflow fail when the tests are incorrect (modify a test case in `internal/greeting/greet_test.go`).

## Build Docker image

1. A `Dockerfile` defining the application image exists in the root directory. To do a container-based deploy we'll use the actions provided by Docker to build the image. Create `.github/workflows/build.yml` with the following content:

    ```yml
    on:
      push

    jobs:
      build:
        steps:
        - uses: actions/checkout@v4
        - name: Set up Docker Buildx
          uses: docker/setup-buildx-action@v3

        - name: Build and push Docker image
          uses: docker/build-push-action@v5
          with:
            push: false
            tags: ghcr.io/${{ github.repository }}:latest
    ```

> [!NOTE]
> 
> The `${{ <expression> }}` syntax is used to access variables, call functions and more. You can read more in [the documentation](https://docs.github.com/en/actions/learn-github-actions/expressions).
>
> In this case, `${{ github.repository }}` is a variable from the [`github` context](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) refers to the owner or and repos, meaning the Docker image will be tagged with `ghcr.io/<user-or-org>/<repo-name>:latest`.

2. Push and verify that the action runs correctly.

3. In order to push the image, we will need to set up [permissions](https://docs.github.com/en/actions/using-jobs/assigning-permissions-to-jobs). With `packages: write` you allow the action to push images to the GitHub Container Registry (GHCR). You can set it at the top-level, for all jobs in the workflow, or for a single job:
    
    ```yml
    jobs:
      build:
        permissions:
          packages: write
        # ... runs-on, steps, etc
    ```

4. We'll have to add a step the `docker/login-action@v3` action to login to GHCR, before we push it. Add the following step before the build and push step:

    ```yml
          - name: Login to GitHub Container Registry
            uses: docker/login-action@v3
            with:
              registry: ghcr.io
              username: ${{ github.actor }}
              password: ${{ github.token }}
    ```

> [!NOTE]
> The `github.token` (often referred to as `GITHUB_TOKEN`) is a special token used to authenticate the workflow job. Read more about it here in [the documentation](https://docs.github.com/en/actions/security-guides/automatic-token-authentication).

5. Finally, modify the build and push step. Set `push: true` to make the action push the image after it's built.

6. Push the changes and make sure the workflow runs successfully. This will push the package to your organization's or your own package registry. The image should be associated with your repo as a package (right-hand side on the main repository page).

## Parallel jobs

Jobs in the same workflow file can be run in parallel if they're not dependent on each other.

1. We also want to lint the app code. In order to get faster feedback, we can lint create a separate job in our `test.yml` workflow file. Create a new job in `test.yml` for linting, which contains the following step:

    ```yml
          - name: Verify formatting
            run: |
              no_unformatted_files="$(gofmt -l $(git ls-files '*.go') | wc -l)"
              exit "$no_unformatted_files"
    ```

    You'll have to checkout the code repository and setup go, similarly to before.

2. Push the code and verify that the workflow runs two jobs successfully.

TODO:
* Require build, tests & linting to succeed to merge PR

## Triggering workflows

Workflows can be triggered in many different ways and can be grouped into four type of events:

* Repository related events
* External events
* Scheduled triggering
* Manual triggering

Repository related events are the most common and are triggered when something happens in the repository. External events and scheduled triggers are not covered in this workshop, but it is nice to know that it is possible. Some example triggers:

```
on: push # Triggers when a push is made to the repository
on: pull_request # Triggers when a pull request is opened or changed
on: workflow_dispatch # Triggers when a user manually requests a workflow to run
```

Some events have filters that can be applied to limit when the workflow should run. For example, the `push`-event has a `branches`-filter that can be used limit the workflow to only run if it is on a specific branch (or branches)

```
on:
  push:
    branches:
      - main
      - 'releases/**'  # Wildcard can be used to limit to a specific set of branches
```

**Tasks**

1. Rewrite the docker build workflow `build.yml` to only be done on main and rewrite the build and lint workflow `test.yml` to only run on PR changes
2. Create a new feature branch, add a new commit with a dummy change (to any file) and finally create a PR to main. Verify that the `test.yml` workflow is run on the feature branch. Merge the PR and verify that the `build.yml`-workflow is only run on the main-branch.
3. Update the `test.yml` workflow and add the event for triggering the workflow manually. Make sure to push the change to main-branch.
4. Go to the [GitHub Actions page of the workflow](/.github/workflows/test.yml) and verify that the workflow can be run manually. A `Run workflow` button should appear to enable you to manually trigger the workflow. 

> [!NOTE]
> In order for the `Run workflow`-button to appear the workflow must exist on the default branch, typically the main-branch

## Reusable workflows

Reusable workflows makes it possible to avoid duplication and reuse common workflow-functionality. They can be [shared within a single repository or by the whole organization](https://docs.github.com/en/actions/using-workflows/reusing-workflows#access-to-reusable-workflows)

To pass information to a shared workflow you should either use the `vars`-context or pass information directly to the workflow.

Reusable workflows are very similar to manual workflows and use the `workflow_dispatch`-trigger. A simple reusable workflow that accepts a config value as input look like this:
``` 
on:
  workflow_call:
    inputs:
      config-value:
        required: true
        type: string
```

### Calling a reusable workflow

To call a reusable workflow in the same repository:
```
jobs:
  call-workflow-passing-data:
    uses: ./.github/workflows/my-reusable-workflow.yml
    with:
      config-value: 'Some value'
```

**Tasks**
1. Create a reusable workflow that runs the test-job specified in `test.yml` and modify `test.yml` to use the reusable workflow for running the tests
2. Create a reusable workflow for the `Build and push Docker image` step in `build.yml` and use a input-parameter to determine if the image should be pushed or not

## Deploying to environment

TODO:
* "Fake deploy" to save time, print name of image to be deployed
* Add environments test, prod in GitHub UI
* Prod should be protected using branch protection rules or rulesets
* Deploying

## Extra: Environment variables and secrets

TODO: Need a use case

## Extra: Reusable composite actions

* Create reusable composite actions for build, use as part of jobs on PR and main pushes


