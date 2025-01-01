import * as vscode from 'vscode';

import { debugChannel } from './logging';

const decorationTypesData: Array<[string, vscode.DecorationRenderOptions]> = [
    ['*', { fontWeight: 'bold' }],
    ['R', { color: 'red' }],
];

const decorationTypes = new Map(decorationTypesData.map(
    ([k, v]) => [k, vscode.window.createTextEditorDecorationType(v)]));

let activeEditor: vscode.TextEditor | undefined;

export function activateStyledown(context: vscode.ExtensionContext) {
    activeEditor = vscode.window.activeTextEditor;

    if (activeEditor) {
        updateDecorations();
    }

    vscode.window.onDidChangeActiveTextEditor((editor) => {
        activeEditor = editor;
        updateDecorations();
    }, null, context.subscriptions);

    vscode.workspace.onDidChangeTextDocument((event) => {
        if (activeEditor && event.document === activeEditor.document) {
            updateDecorations();
        }
    }, null, context.subscriptions);
}

function updateDecorations() {
    if (!activeEditor || activeEditor.document.languageId !== 'styledown') {
        return;
    }
    debugChannel.appendLine('updateDecorations start');
    const decorations = new Map<string, vscode.Range[]>();
    for (let i = 1; i < activeEditor.document.lineCount; i += 2) {
        const styleLine = activeEditor.document.lineAt(i).text;
        debugChannel.appendLine(`style line ${i}: ${styleLine}`)
        for (let j = 0; j < styleLine.length; j++) {
            const styleChar = styleLine[j];
            if (!decorationTypes.has(styleChar)) {
                continue;
            }
            if (!decorations.has(styleChar)) {
                decorations.set(styleChar, []);
            }
            decorations.get(styleChar)!.push(new vscode.Range(i - 1, j, i - 1, j + 1));
        }
    }
    debugChannel.appendLine(`updateDecorations: ${JSON.stringify(Object.fromEntries(decorations))}`);
    for (const [styleChar, ranges] of decorations.entries()) {
        activeEditor.setDecorations(decorationTypes.get(styleChar)!, ranges);
    }
}
