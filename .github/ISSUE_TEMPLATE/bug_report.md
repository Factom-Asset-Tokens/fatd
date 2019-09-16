---
name: Bug report
about: Create a bug report to help us improve
title: "[BUG]"
labels: bug
assignees: ''

---

!!! **DO NOT REPORT SECURITY ISSUES HERE** !!!

If the issue is security related, copy this template and email the issue to
[adam@canonical-ledgers.com](mailto:adam@canonical-ledgers.com).

!!! **DO NOT REPORT SECURITY ISSUES HERE** !!!

**Describe the bug**
A clear and concise description of what the bug is.

**Describe the environment (please complete the following)**
- OS: [e.g. Windows 10, Max OS X, Linux 5.2.13]
- Factom Network: [e.g. mainnet, testnet, or custom]
- `fatd` State: [e.g. starting up, syncing, synced, shutting down]
- `fatd` Version: [e.g. v0.6.0.r54.ge7a7ca1]
- `fat-cli` Version: [Will likely be the same as `fatd`'s version, but not necessarily]

NOTE: The `fatd` version is printed on the first INFO line, which is the most
reliable way to retrieve the version after a crash. If the first line was lost,
simply re-run the same binary. Use `fat-cli --version` for the `fat-cli`
version and to retrieve the `fatd` version from a running instance.

**To Reproduce**
Steps to reproduce the behavior:
1. Start `fatd` with following flags/environment variables ...
2. Wait for sync ...
3. Use `fat-cli` to query for transaction ...
4. See error ...

**Expected behavior**
A clear and concise description of what you expected to happen.

**Output**
If applicable, please attach as much relevant `fatd` or `fat-cli` output as
possible.

**Additional context**
Add any other context about the problem here.
