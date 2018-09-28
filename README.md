# lsp-filter

lsp-filter can wrap a language server binary and customize the `initialize`
message response by disabling certain semantic providers.

For example, you can use cquery and clangd together by disabling completion
for cquery and disabling everything except completion for clangd.

## Installation

```sh
$ go get -u github.com/jacobdufault/lsp-filter
$ lsp-filter # prints out help/usage
```

## Example

Install the vscode clangd and cquery extensions, and use `clangd-chromium.sh`
and `cquery-chromium.sh` as the language server binary instead of `cquery` or
`clangd`.

```sh
$ cat clangd-chromium.sh
#!/usr/bin/env /bin/sh
exec lsp-filter clangd enable completion signatureHelp codeAction executeCommand -- "$@"

$ cat cquery-chromium.sh
#!/usr/bin/env /bin/sh
exec lsp-filter cquery disable completion signatureHelp codeAction -- "$@"
```

vscode settings file:
```js
{
  // ...
  "clangd.path": "clangd-chromium.sh",
  "cquery.launch.command": "cquery-chromium.sh",
  // ...

  // You'll probably also want to turn off diagnostics for cquery, since clangd
  // provides those.
  "cquery.diagnostics.onParse": false,
  "cquery.diagnostics.onType": false,
}
```

Make sure `clangd-chromium.sh` and `cquery-chromium.sh` are in your `PATH`.
