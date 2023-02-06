
<!--
Thank you for opening a pull request. Please provide:

- A clear summary of your changes

- Descriptive and succinct commit messages with the format:
  """
  [topic]: [short description]

  [Longer description]

  Signed-off-by: [Your Name] <[your email]>
  """

  Topic will generally be the go ceph package dir you are working in.

- Ensure checklist items listed below are accounted for
-->

## Checklist
- [ ] Added tests for features and functional changes
- [ ] Public functions and types are documented
- [ ] Standard formatting is applied to Go code
- [ ] Is this a new API? Added a new file that begins with `//go:build ceph_preview`
- [ ] Ran `make api-update` to record new APIs

New or infrequent contributors may want to review the go-ceph [Developer's Guide](https://github.com/ceph/go-ceph/blob/master/docs/development.md) including the section on how we track [API Status](https://github.com/ceph/go-ceph/blob/master/docs/development.md#api-status) and the [API Stability Plan](https://github.com/ceph/go-ceph/blob/master/docs/api-stability.md).

The go-ceph project uses mergify. View the [mergify command guide](https://docs.mergify.com/commands/#commands) for information on how to interact with mergify. Add a comment with `@Mergifyio` `rebase` to rebase your PR when github indicates that the PR is out of date with the base branch.
