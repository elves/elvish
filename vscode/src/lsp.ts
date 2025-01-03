/** Hooks up LSP for Elvish sources. */

import * as vscode from 'vscode';
import { LanguageClient } from 'vscode-languageclient/node';

export function activateLsp(context: vscode.ExtensionContext) {
    const client = new LanguageClient(
        "elvish",
        "Elvish Language Server",
        { command: "elvish", args: ["-lsp"] },
        { documentSelector: [{ scheme: "file", language: "elvish" }] }
    );
    client.start();

    return () => { client.stop() };
}
