// Copyright 2018 Jacob Dufault
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func panicIfError(e error) {
	if e != nil {
		panic(e)
	}
}

func printHelp() {
	fmt.Printf(`%[1]v <bin> <mode> <providers...> -- <args...>

bin: the binary of the language server

mode: enable or disable
    enable: allow only the specified providers
    disable: allow all providers except the specified ones

providers: language server capability without the "Provider" at the end
    codeAction codeLens completion definition documentFormatting
    documentHighlight documentRangeFormatting documentLink documentSymbol
    hover implementation references rename signatureHelp typeDefinition
    workspaceSymbol

args: arguments to pass to the language server binary <bin>

Examples:
  %[1]v cquery disable completion codeAction --
  %[1]v clangd enable completion codeAction --
`, os.Args[0])
	os.Exit(1)
}

func indexOf(element string, data []string) int {
	for i, v := range data {
		if element == v {
			return i
		}
	}
	return -1
}

func contains(element string, data []string) bool {
	return indexOf(element, data) >= 0
}

type mode int

const (
	modeEnable mode = iota
	modeDisable
)

func parseMode(value string) mode {
	if value == "enable" {
		return modeEnable
	}
	if value == "disable" {
		return modeDisable
	}
	fmt.Fprintf(os.Stderr, "Mode must be either enable or disable, not %v\n", value)
	os.Exit(1)
	return modeDisable
}

func main() {
	if len(os.Args) < 4 {
		printHelp()
	}
	ourArgEnd := indexOf("--", os.Args)
	if ourArgEnd < 0 {
		printHelp()
	}

	binary := os.Args[1]
	mode := parseMode(os.Args[2])

	providers := os.Args[3:ourArgEnd]
	appOpts := os.Args[ourArgEnd+1:]
	fmt.Fprintf(os.Stderr, "Running binary %v %v in mode %v %v\n", binary, appOpts, os.Args[2], providers)

	cmd := exec.Command(binary, appOpts...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	lsStdout, e := cmd.StdoutPipe()
	panicIfError(e)
	e = cmd.Start()
	panicIfError(e)

	go stdoutReader(lsStdout, mode, providers)

	cmd.Wait()
	panicIfError(e)
}

func stdoutWrite(b []byte) {
	osStdout := bufio.NewWriter(os.Stdout)
	_, e := fmt.Fprintf(osStdout, "Content-Length: %d\r\n\r\n", len(b))
	panicIfError(e)
	_, e = osStdout.Write(b)
	osStdout.Flush()

	osStderr := bufio.NewWriter(os.Stderr)
	fmt.Fprintf(osStderr, "%v wrote the following to stderr:\n\t", os.Args[0])
	osStderr.Write(b)
	osStderr.Flush()

	panicIfError(e)
}

func stdoutReader(lsStdout io.ReadCloser, mode mode, providers []string) {
	// Build scanner which will process LSP messages.
	scanner := bufio.NewScanner(lsStdout)
	scanner.Split(jsonRpcSplitFunc)
	osStdout := bufio.NewWriter(os.Stdout)
	for scanner.Scan() {
		fmt.Fprintln(os.Stderr, scanner.Text())
		var f interface{}
		err := json.Unmarshal(scanner.Bytes(), &f)
		panicIfError(err)
		jsonRoot, had := f.(map[string]interface{})
		if !had {
			stdoutWrite(scanner.Bytes())
			continue
		}
		jsonResult, had := jsonRoot["result"].(map[string]interface{})
		if !had {
			stdoutWrite(scanner.Bytes())
			continue
		}
		jsonCapabilities, had := jsonResult["capabilities"].(map[string]interface{})
		if !had {
			stdoutWrite(scanner.Bytes())
			continue
		}

		if mode == modeEnable {
			// Only enable providers in the whitelist
			for k := range jsonCapabilities {
				if !strings.HasSuffix(k, "Provider") {
					continue
				}
				rawName := strings.TrimSuffix(k, "Provider")
				if !contains(rawName, providers) {
					jsonCapabilities[k] = false
				}
			}
		} else if mode == modeDisable {
			// Only disable the specified providers
			for k := range jsonCapabilities {
				if !strings.HasSuffix(k, "Provider") {
					continue
				}
				rawName := strings.TrimSuffix(k, "Provider")
				if contains(rawName, providers) {
					jsonCapabilities[k] = false
				}
			}
		}

		// Serialize and write message
		b, e := json.Marshal(f)
		panicIfError(e)
		stdoutWrite(b)

		break
	}
	panicIfError(scanner.Err())

	// Write the rest of the content blindly
	var buffer [1024]byte
	for {
		n, e := lsStdout.Read(buffer[:])
		if e != nil {
			break
		}

		w := bufio.NewWriter(os.Stderr)
		w.Write(buffer[:n])
		w.Flush()

		_, e = osStdout.Write(buffer[:n])
		osStdout.Flush()
		if e != nil {
			break
		}
	}

	// Close stdout.
	panicIfError(lsStdout.Close())
}
