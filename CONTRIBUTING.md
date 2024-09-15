# Contributing to Athena

The Blocsoc team maintains guidelines for contributing to the Blocsoc repos. Check out our [website](https://blocsoc.iitr.ac.in/) if you want to learn more about us.

Also, give a read to our [code of conduct](./CODE_OF_CONDUCT.md) if you haven't already.

## Bugs and Feature Request

Before you make your changes, check to see if an [issue](https://github.com/BlocSoc-iitr/Athena/issues) exists already for the change you want to make.

### Opening issues

If you spot something new or want to change something, please make sure there isn't already an [issue](https://github.com/BlocSoc-iitr/Athena/issues) related to it.

If no issue addresses your problem, please open a new one with an accurate description of the problem. Please add, if possible, labels to have an overview of what they are targetting, and having an easier time filtering them out.

### Making Pull Requests

When you're done making changes and you'd like to propose them for review through a Pull Request.
Please make sure the PR mentions the related issue, detailing the changes made and potential side effects.

If your PR is not ready for review and merge because you are still working on it, please convert it to a draft.

## DOs and DON'Ts

Please do:

* **DO** give priority to the current style of the project or file you're changing even if it diverges from the general guidelines.
* **DO** include tests when adding new features. When fixing bugs, start with adding a test that highlights how the current behavior is broken.

Please do not:

* **DON'T** copy paste code from other data decoder implementations without a good argument to back it up.
* **DON'T** surprise us with big pull requests. Instead, file an issue and start a discussion so we can agree on a direction before you invest a large amount of time.

## Setting Up Git Hooks

To ensure code quality and consistency, this project uses a custom pre-commit hook. Follow these steps to set up the pre-commit hook:

Run the setup script to install the pre-commit hook:

```bash
chmod +x ./setup-hooks.sh
./setup-hooks.sh
```

This script will copy the pre-commit hook from the githooks directory to your local .git/hooks directory.

Verify that the pre-commit hook is installed and executable by running:

```bash
ls -l .git/hooks/pre-commit
```

You should see that the pre-commit file is listed and has executable permissions.
