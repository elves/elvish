import * as path from 'path';
import * as child_process from 'child_process';
import * as vscode from 'vscode';
import { LanguageClient } from 'vscode-languageclient/node';

let client: LanguageClient | undefined;

export function activate(context: vscode.ExtensionContext) {
    client = new LanguageClient(
        "elvish",
        "Elvish Language Server",
        { command: "elvish", args: ["-lsp"] },
        { documentSelector: [{ scheme: "file", language: "elvish" }] }
    );
    client.start();

    context.subscriptions.push(vscode.commands.registerCommand(
        'elvish.updateTranscriptOutputForCodeAtCursor',
        updateTranscriptOutputForCodeAtCursor));
}

export function deactivate() {
    return client?.stop();
}

interface UpdateInstruction {
    fromLine: number;
    toLine: number;
    content: string;
}

async function updateTranscriptOutputForCodeAtCursor() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        return;
    }
    const {dir, base} = path.parse(editor.document.uri.fsPath);
    // VS Code's line number is 0-based, but the ELVISH_TRANSCRIPT_RUN protocol
    // uses 1-based line numbers. This is also used in the UI, where the user
    // expects 1-based line numbers.
    const lineno = editor.selection.active.line + 1;

    await vscode.window.withProgress({
        location: vscode.ProgressLocation.Notification,
        title: `Running ${base}:${lineno}...`
    }, async (progress, token) => {
        // Transcript tests uses what's on the disk, so we have to save the
        // document first.
        await editor.document.save();

        // See godoc of pkg/eval/evaltest for the protocol.
        const {error, stdout} = await exec(
            "go test -run TestTranscripts",
            {
                cwd: dir,
                env: {...process.env, ELVISH_TRANSCRIPT_RUN: `${base}:${lineno}`},
            });
        if (error) {
            const match = stdout.match(/UPDATE (.*)$/m);
            if (match) {
                const {fromLine, toLine, content} = JSON.parse(match[1]) as UpdateInstruction;
                const range = new vscode.Range(
                    new vscode.Position(fromLine-1, 0), new vscode.Position(toLine-1, 0));
                editor.edit((editBuilder) => {
                    editBuilder.replace(range, content);
                });
            } else {
                vscode.window.showWarningMessage(`Unexpected test failure: ${stdout}`)
            }
        } else {
            // TODO: Distinguish two different cases:
            //
            // - Output is already up-to-date
            // - Cursor is in an invalid position.
            //
            // This needs to be detected by evaltest first.
            vscode.window.showInformationMessage('Nothing to do.')
        }
    });
}

// Wraps child_process.exec to return a promise.
function exec(cmd: string, options: child_process.ExecOptions):
        Promise<{error: child_process.ExecException|null, stdout: string, stderr: string}> {
    return new Promise((resolve) => {
        child_process.exec(cmd, options, (error, stdout, stderr) => {
            resolve({error, stdout, stderr});
        });
    });
}
