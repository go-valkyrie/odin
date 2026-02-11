# Contributing to Odin

Thank you for your interest in contributing to Odin! This document outlines our guidelines and expectations for
contributions.

## AI Usage Policy

### Overview

This project uses AI tools as part of its development process. Starting from commit
`05913d872188289ea923c825b19bc17a971b3448`, AI assistance has been used under the guidelines described below.

**All use of AI tools in repository contributions must be disclosed.** This policy applies to code, tests,
documentation, examples, and any other content committed to the repository. It does not apply to how you use AI tools
to understand or use Odin itself—only to what you submit as contributions.

### Why We Require Disclosure

**Transparency and trust are fundamental to open source.** There is well-deserved skepticism about AI tools in
creative fields, including software engineering. For those who choose to use Odin, it's important that we're clear
about how we use these tools and what we expect from contributors. This transparency demonstrates that we take quality
seriously, that we care about copyright (including our own—open source only works because copyright law gives us a
legal foundation when someone violates the license), and that we use tools responsibly.

We require disclosure for three specific reasons:

1. **Quality Signal**: Non-disclosed AI usage in low-quality contributions will result in immediate rejection. We need
   to trust that contributors who don't disclose AI usage genuinely wrote the code themselves.

2. **Copyright and Authorship**: AI-generated content without substantial human involvement may not be eligible for
   copyright protection. Since Odin is MIT-licensed, contributions must be legally capable of being licensed. Simply
   prompting an AI and submitting the output does not constitute sufficient human authorship.

3. **Derivative Works**: While less of a concern for this project, we need human review to ensure AI outputs don't
   inadvertently include problematic content from training data.

### What Constitutes Acceptable AI-Assisted Contributions

AI tools are valuable when used as tools in an iterative, creative process. Acceptable use involves:

- **Creative Input**: You make architectural decisions, design choices, and guide the implementation approach
- **Iterative Development**: You review AI outputs, request changes, refactor code, and ensure it meets best practices
- **Understanding**: You can explain every part of the contribution and defend the decisions made
- **Human Decision-Making**: You are actively involved in the creative process, not just running prompts

**What matters is not what percentage of keystrokes came from you versus the AI, but that you were actively engaged in
creative decision-making throughout the process.**

### What Will Be Rejected

**"Prompt and submit" contributions will be rejected immediately.** This means:

- Giving an AI a feature request and submitting the output without meaningful review or modification
- Contributions that appear to be unreviewed AI output, regardless of disclosure
- Any contribution that looks like "AI slop" will be rejected, even if you claim no AI was used (in which case, we'll
  assume you weren't honest about AI usage)

For first-time contributors especially: be explicit about your process. If something appears suspicious, you may be
asked to explain your contributions. Obvious low-quality AI-generated PRs will be closed without comment.

### How to Disclose AI Usage

When committing AI-assisted work, add the AI tool as a co-author in your commit message:

```
feat: add new feature

Implementation details here.

Co-authored-by: Claude Sonnet 4.5 <noreply@anthropic.com>
```

Or for other tools:
```
Co-authored-by: GitHub Copilot <noreply@github.com>
```

This simple footer is sufficient disclosure.

### Edge Cases

**Test Fixtures and Mock Data**: AI-generated test data is acceptable, but you must verify it makes sense. Common
problems include: AI changing test data to make broken tests pass, only covering happy paths, or generating fixtures
that match a flawed implementation. Review this carefully.

**Mechanical Refactoring**: Using AI for tedious mechanical refactoring (renaming, reformatting, moving code) is fine,
as long as:
- The change itself makes sense
- Only refactoring is present in that commit (don't mix refactoring with feature work)

**Commit Messages and PR Descriptions**: These must be reviewed by a human like any other content. They need to
accurately reflect the change and be concise. Overly verbose AI-generated commit messages are a common issue.

## Developer Certificate of Origin (DCO)

This project uses the Developer Certificate of Origin (DCO) to ensure that contributors have the legal right to submit
their contributions. This is especially important given our AI usage policy and copyright considerations.

By contributing to this project, you certify that:

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

### How to Sign Off Your Commits

To sign off your commits, add the `-s` flag when committing:

```bash
git commit -s -m "Your commit message"
```

This adds a `Signed-off-by` line to your commit message:

```
Your commit message

Signed-off-by: Your Name <your.email@example.com>
```

**All commits must include this sign-off.** Pull requests with unsigned commits will not be accepted.

If you're using AI assistance, your commit will include both the DCO sign-off and the AI co-author attribution:

```
feat: add new feature

Implementation details here.

Signed-off-by: Your Name <your.email@example.com>
Co-authored-by: Claude Sonnet 4.5 <noreply@anthropic.com>
```

By signing off, you're certifying that you have the right to submit your contribution—which requires the substantial
human involvement described in our AI usage policy.

## Commit Message Format

This project uses the [Conventional Commits](https://www.conventionalcommits.org/) format for commit messages. This
provides a consistent structure that makes the project history easier to understand and enables automated tooling.

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

Use one of the following types:

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, whitespace, etc.)
- **refactor**: Code changes that neither fix a bug nor add a feature
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Changes to build process, tooling, dependencies, etc.

### Scope

Optional. Indicates what part of the codebase is affected:

```
feat(template): add support for custom resources
fix(docs): correct schema expansion logic
```

### Breaking Changes

Mark breaking changes with `!` after the type/scope:

```
feat!: change bundle configuration format
```

Also include `BREAKING CHANGE:` in the commit body or footer with details about the change and migration path.

### Examples

**Simple feature:**
```
feat: add odin test command

Signed-off-by: Your Name <your.email@example.com>
```

**Bug fix with scope:**
```
fix(schema): handle nil values in walk function

Signed-off-by: Your Name <your.email@example.com>
```

**Feature with AI assistance:**
```
feat: add bundle validation

Implements validation for component references and resource
naming conflicts.

Signed-off-by: Your Name <your.email@example.com>
Co-authored-by: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Breaking change:**
```
feat!: change odin.toml registry format

BREAKING CHANGE: Registry configuration now uses `[[registries]]`
instead of `[[registry]]`. Update your odin.toml files:

Before:
[[registry]]
prefix = "example.com"
host = "registry.example.com"

After:
[[registries]]
module-prefix = "example.com"
registry = "registry.example.com"

Signed-off-by: Your Name <your.email@example.com>
```

### Guidelines

- Keep the description line under 72 characters
- Use imperative mood ("add feature" not "added feature")
- Don't end the description with a period
- Use the body to explain *what* and *why*, not *how*
- Reference issues in the footer: `Fixes #123` or `Closes #456`

## Types of Contributions

We welcome various types of contributions beyond just code:

### Bug Reports

If you find a bug, please open a GitHub issue with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version, Odin version)
- Minimal example bundle if applicable

### Feature Requests

Have an idea for a new feature? Open a GitHub issue describing:
- The use case and problem you're trying to solve
- Proposed solution or approach
- Why this would be valuable to other users

### Documentation

Documentation improvements are always welcome:
- Fixing typos or unclear explanations
- Adding examples
- Improving error messages
- Writing guides or tutorials

Documentation changes follow the same AI usage and DCO policies as code.

### Examples and Templates

Contributions of example bundles or component templates help others learn:
- Real-world usage examples
- Best practice demonstrations
- Common patterns and solutions

### Code Contributions

For code changes:
- Bug fixes: Reference the issue in your PR
- Features: Discuss in an issue first for larger changes
- Refactoring: Explain the benefits and keep changes focused

All code contributions must follow the AI usage policy, DCO requirements, and commit message format described above.

## Communication

- **Bug reports**: Open a [GitHub issue](https://github.com/go-valkyrie/odin/issues)
- **Feature requests**: Open a [GitHub issue](https://github.com/go-valkyrie/odin/issues)
- **Questions and support**: Use [GitHub Discussions](https://github.com/go-valkyrie/odin/discussions)
- **Security issues**: Do not open public issues; see SECURITY.md (if you have a security policy) or contact the
  maintainers directly

Use issues for actionable items (bugs, features). Use discussions for "How do I...?" questions, general feedback, or
open-ended conversations.

## General Contribution Guidelines

- Follow existing code patterns and style in the codebase
- Include tests for new functionality
- Update documentation as needed
- Use conventional commit format for all commits
- Sign off all commits with `git commit -s`
- Be prepared to explain and defend your design decisions

## Questions?

If you're unsure whether your use of AI tools meets these guidelines, ask before submitting. We're happy to clarify.

For general questions or feedback about Odin, please open an issue.

---

**Note**: This CONTRIBUTING.md file was written with AI assistance, following the guidelines described above.
