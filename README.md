# lsp-filter

lsp-filter can wrap a language server binary and customize the `initialize`
message response by disabling certain semantic providers.

For example, you can use cquery and clangd together by disabling completion for
cquery and disabling everything except completion for clangd.

## Installation

```sh
$ go get -u github.com/jacobdufault/lsp-filter
$ lsp-filter # prints out help/usage
```

## Example

Install the vscode clangd and cquery extensions, and use `clangd-filter.sh` and
`cquery-filter.sh` as the language server binary instead of `cquery` or
`clangd`.

```sh
$ cat clangd-filter.sh
#!/usr/bin/env /bin/sh
exec lsp-filter clangd enable completion signatureHelp codeAction executeCommand -- "$@"

$ cat cquery-filter.sh
#!/usr/bin/env /bin/sh
exec lsp-filter cquery disable completion signatureHelp codeAction -- "$@"
```

vscode settings file:

```js
{
  // ...
  "clangd.path": "clangd-filter.sh",
  "cquery.launch.command": "cquery-filter.sh",
  // ...

  // You'll probably also want to turn off diagnostics for cquery, since clangd
  // provides those.
  "cquery.diagnostics.onParse": false,
  "cquery.diagnostics.onType": false,
}
```

Make sure `clangd-filter.sh` and `cquery-filter.sh` are in your `PATH`.

### Custom compile_commands.json directory

By default clangd only searches for compile_commands.json in the parent
directories of the source files that are being analyzed. However, some build
systems such as CMake store their generated compile_commands.json in the build
directory itself which is often not a parent directory of the source files (e.g.
when using src/ and build/ directories). To make clangd look for the
compile_commands.json file in a custom directory we have to specify an extra
option when launching clangd:

```sh
$ cat clangd-filter.sh
#!/usr/bin/env /bin/sh
exec lsp-filter clangd enable completion signatureHelp codeAction executeCommand -- -compile-commands-dir=$PWD/build "$@"
```

This assumes that clangd is started in the root directory of the current
project so that `$PWD` refers to the root directory of the project.
