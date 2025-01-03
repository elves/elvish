/** Implements a command to update command output in Elvish transcripts. */

import * as path from 'path';
import * as child_process from 'child_process';
import * as vscode from 'vscode';

export function activateTranscript(context: vscode.ExtensionContext) {
    context.subscriptions.push(vscode.commands.registerCommand(
        'elvish.updateTranscriptOutputForCodeAtCursor',
        updateTranscriptOutputForCodeAtCursor));
    context.subscriptions.push(vscode.commands.registerCommand(
        'elvish.openTranscriptPromptBelow', openTranscriptPromptBelow));
    return () => { };
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
    const { dir, base } = path.parse(editor.document.uri.fsPath);
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
        const { error, stdout } = await exec(
            "go test -run TestTranscripts",
            {
                cwd: dir,
                env: { ...process.env, ELVISH_TRANSCRIPT_RUN: `${base}:${lineno}` },
            });
        if (error) {
            const match = stdout.match(/UPDATE (.*)$/m);
            if (match) {
                const { fromLine, toLine, content } = JSON.parse(match[1]) as UpdateInstruction;
                const range = new vscode.Range(
                    new vscode.Position(fromLine - 1, 0), new vscode.Position(toLine - 1, 0));
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
    Promise<{ error: child_process.ExecException | null, stdout: string, stderr: string }> {
    return new Promise((resolve) => {
        child_process.exec(cmd, options, (error, stdout, stderr) => {
            resolve({ error, stdout, stderr });
        });
    });
}

const headingPattern = /^(#{1,3}) .* \1$/;
const promptPattern = /^[~/][^ ]*> /;

async function openTranscriptPromptBelow() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        return;
    }
    const document = editor.document;

    const cursorLine = editor.selection.active.line;
    // Find the current prompt.
    let prompt = '~> ';
    let i;
    for (i = cursorLine; i >= 0; i--) {
        const line = document.lineAt(i).text;
        const match = line.match(promptPattern);
        if (match) {
            prompt = match[0];
            break;
        }
        if (line.match(headingPattern)) {
            break;
        }
    }

    // Find the correct line to insert.
    for (i = cursorLine + 1; i < document.lineCount; i++) {
        const line = document.lineAt(i).text;
        if (line.match(headingPattern) || line.match(promptPattern)) {
            break;
        }
    }
    for (; i > cursorLine && document.lineAt(i - 1).text === ''; i--) {
    }

    // Now insert and move the cursor.
    editor.edit((editBuilder) => {
        editBuilder.insert(new vscode.Position(i, 0), prompt + "\n");
    });
    const newCursorPos = new vscode.Position(i, prompt.length);
    editor.selection = new vscode.Selection(newCursorPos, newCursorPos);
}
