# rancher

`rancher` is a utility for creating Git branches. Very simple, run `rancher`, fill out the fields,
and get a git branch at the end of it. Branches are created with the format `{feature type?}/{ticket?}/{description}`.

> [!NOTE]
> It's possible to customise the separator, but not yet the layout of the components in the branch.
> If this is a feature you really want, make an issue!

![Rancher demo](demo.gif)

Rancher also supports a basic jira integration. Create an API token
[here](https://id.atlassian.com/manage-profile/security/api-tokens) and save it to the rancher
configuration file at `$HOME/.config/rancher/rancher.yml`.

```!yaml
jira:
    username: <YOUR EMAIL ADDRESS>
    api-token: <YOUR API TOKEN>
```

Subsequent runs of `$ rancher -jira` will populate the ticket list with your assigned tickets in
active sprints. Selecting one of these will fill out the `{ticket}` and `{description}` fields of
the branch.

### Installation

Install this with `$ go install github.com/mr55p-dev/rancher@latest`.

### Contributing

Feel free to raise a PR with any issues or features. For requests, make an issue!
