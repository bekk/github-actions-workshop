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

Fork the repository to your personal account using the GIT CLI with `gh repo fork --clone` or by using the GitHub UI. Continue reading the workshop README on the forked repo since some of the links will only work there. 

This repository contains a simple Go app. You do not need to know Go, nor use any Golang tooling. We will, unless explicitly specified otherwise, only modify files in the special `.github/` directory.

Start by [enabling GitHub Actions in your fork](../../actions)

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
        runs-on: 'ubuntu-latest'
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

1. We also want to lint the app code. In order to get faster feedback, we can create a separate lint-job in our `test.yml` workflow file. Create a new job in `test.yml` for linting, which contains the following step:

    ```yml
          - name: Verify formatting
            run: |
              no_unformatted_files="$(gofmt -l $(git ls-files '*.go') | wc -l)"
              exit "$no_unformatted_files"
    ```

    You'll have to checkout the code repository and setup go, similarly to before.

2. Push the code and verify that the workflow runs two jobs successfully.

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

1. Rewrite the docker build workflow `build.yml` to only be done on main and rewrite the build and lint workflow `test.yml` to only run on PR changes. Push the changes to main-branch and observe that only the build-workflow is executed.
2. Create a new feature branch, add a new commit with a dummy change (to any file) and finally create a PR to main. Verify that the `test.yml` workflow is run on the feature branch. Merge the PR and verify that the `build.yml`-workflow is only run on the main-branch.
3. Update the `test.yml` workflow and add the event for triggering the workflow manually. Make sure to push the change to main-branch.
4. Go to the [GitHub Actions page of the workflow](/../../actions/workflows/test.yml) and verify that the workflow can be run manually. A `Run workflow` button should appear to enable you to manually trigger the workflow. 

> [!NOTE]
> In order for the `Run workflow`-button to appear the workflow must exist on the default branch, typically the main-branch

## Reusable workflows

Reusable workflows makes it possible to avoid duplication and reuse common workflow-functionality. They can be [shared within a single repository or by the whole organization](https://docs.github.com/en/actions/using-workflows/reusing-workflows#access-to-reusable-workflows)

To pass information to a shared workflow you should either use [the `vars`-context](https://docs.github.com/en/actions/learn-github-actions/contexts#about-contexts) or pass information directly to the workflow. The variables for the `vars`-context can be [found here](../../github-actions-workshop/settings/variables/actions).

Reusable workflows use the `workflow_call`-trigger. A simple reusable workflow that accepts a config value as input look like this:
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

1. Create a reusable workflow that runs the test-job specified in `test.yml` and modify `test.yml` to use the reusable workflow for running the tests
2. Create a reusable workflow for the `Build and push Docker image` step in `build.yml` and use a input-parameter to determine if the image should be pushed or not

## Deploying to environment

For the purposes of this workshop, we'll not actually deploy to any environment, but create a couple of GitHub environments to demonstrate how it would work. You can use environments to track deploys to a given environment, and set environment-specific variables required to deploy your application.

Example of a deployment job to an environment:
```
jobs:
  deployment:
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: deploy
```

1. Navigate to [Settings > Environments](../../settings/environments) and create two new environments: `test` and `production`. For each environment set a unique environment variable, `WORKSHOP_ENV_VARIABLE`.

2. Create a new workflow in `.github/workflows/deploy.yml`. This workflow should trigger on `workflow_dispatch`, and take three inputs: `environment` of type `environment`, and the strings `imageName` and `digest`. It should have a single job, `deploy`, and here it should just "fake" the deploy by printing the `imageName` and `digest`. All inputs should be required [set `required` to `true`](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_dispatchinputsinput_idrequired). The job should run in context of the input environment and print the environment variable `${{ vars.WORKSHOP_ENV_VARIABLE }}` configured for the environment.

3. Push the new workflow (to main-branch), and verify that you get a dropdown to select the environment when you trigger it, and that the value of `WORKSHOP_ENV_VARIABLE` is printed for the chosen environment.

## Job dependencies

Jobs can depend on each other. We'll now create a workflow that builds, then deploys the docker image to test and production, in that order.

1. Modify the `deploy.yml` to make it reusable by adding a `workflow_call` trigger. It should have the same inputs as the `workflow_dispatch` trigger.

2. Modify your reusable build action to propagate outputs. You'll need to add an `id: build-push` to the step that builds the image. Then, you can add an `outputs` object property to the job and the workflow.

    To get the correct outputs from the `docker/build-push-action` action, you should use `fromJson(jobs.build.outputs.metadata)['image.name']` and `jobs.build.outputsdigest` outputs from the build-push action as `imageName` and `digest` respectively. You can read more about the `fromJson` expression in [the documentation](https://docs.github.com/en/actions/learn-github-actions/expressions#fromjson).

    Take a look at [the documentation](https://docs.github.com/en/actions/using-workflows/reusing-workflows#using-outputs-from-a-reusable-workflow) for a complete example of outputs for a reusable workflow.

3. Expand your (non-reusable) build workflow with a couple of more jobs: `deploy-test` and `deploy-production`. These jobs should reuse the `deploy.yml` workflow, use `imageName` and `digest` outputs from the `build` job and use correct environments. You have to specify `needs` for the deploy jobs, take a look at [the `needs` context and corresponding example](https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context).

4. Push the workflow, and verify that the jobs run correctly, printing the correct docker image specification and environment variable. The `deploy-test` job should also finish before the `production-test` job starts.

## Branch protection rules

Many teams want require code reviews, avoid accidental changes, run tests or ensure that the formatting is correct on all new code before merging. These restrictions can be done using [*branch protection rules*](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches).

You can find branch protections rules by going to [Settings > Branches](../../settings/branches) (requires repository administrator privileges). Let's create a branch protection rule for the `main` branch:

1. Set `main` as the branch name pattern.

2. Set the setting "Require a pull request before merging", and untick the "Require Approvals" sub-setting.

3. Set the setting "Require status checks to pass before merging", and make sure that both the jobs for linting and testing are selected.

4. Set the "Do not allow bypassing the above settings" setting to disallow administrator overrides, and finally click "Create".

5. Create a change (e.g. text change in the README). Try pushing the change from your local computer directly to `main` and verify that it gets rejected (if you're using the GitHub UI, you will be forced to create a branch).

6. Create a change on a separate branch, push it and create a PR. Verify that you cannot merge it until the status checks have passed.

7. Optionally, turn off the "Require a pull request before merging" and/or "Do not allow bypassing the above settings" settings before you continue, to simplify the rest of the workshop. Read through the list of settings once more and research options you want to know more about in [the documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches).

> [!TIP]
> Branch protection rules will disallow force pushes for everyone, including administrators, by default, but this can be turned on again in the settings.

## Extra: Reusable composite actions

* Create reusable composite actions for build, use as part of jobs on PR and main pushes

## Other extras:


* Gated prod deploy
* Don't trigger build on non-source code changes
* Only deploy prod on main branch
* Environment secrets
