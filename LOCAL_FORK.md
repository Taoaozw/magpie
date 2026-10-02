# Local fork

This fork starts from upstream `v0.1.661` (commit
`f1e265ea3d9025bbb38e8e5c8e6b285d6d51a8fc`). Its changes live on
`clive/pi-messages`.

## Pi protocol selection

Pi and Omo use `anthropic-messages` for Claude Code subscription models.
Generated entries point to the gateway root, allowing the Anthropic SDK to
append `/v1/messages`. The existing adaptive-thinking and effort mapping is
used. Selecting a model or refreshing the catalog regenerates these same
entries, so manual protocol overrides are no longer necessary.

This is a client protocol preference. `Provider.Native()` and the gateway's
`native_endpoints` retain their meaning of HTTP pass-through capability.
Claude Code subscription requests still use the CLI bridge; Messages preserves
more information at the client boundary but does not make that bridge lossless.
Routing groups keep the provider's Chat default because they can choose models
with different upstream protocols. Pencil retains its existing Chat behavior.
Codex subscription models retain `openai-responses`; the Codex CLI's gateway
provider also retains `wire_api = "responses"`.

## Verification and build

```sh
go test ./internal/agent ./internal/provider -run 'TestPi|TestNative' -count=1
make test
make app
codesign --force --deep --sign - magpie.app
```

Keep the default version produced by `git describe`, for example
`v0.1.661-1-g<commit>`. Magpie's updater recognizes this as a source build and
does not replace it with an official release. Do not override `VERSION` with a
plain release version when installing this fork locally.

Before replacing the installed application, back up the complete app bundle
and the local configuration files. Stop the old app, copy the built bundle to
`/Applications/magpie.app`, and start it again. Keep credential/configuration
backups outside this repository.
