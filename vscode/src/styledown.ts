/**
 * Implements styling and folding support for Styledown sources.
 *
 * Limitations:
 *
 * - Only a small subset of custom styling is supported.
 * - Multi-width characters are not handled correctly (matching is based on
 *   UTF-16 characters).
 * - The cursor indicator extension used in TTY tests is not implemented.
 */

import * as vscode from 'vscode';

import { applyStyling } from './utils/styling';

const decorationTypeForStyling = new Map<string, vscode.TextEditorDecorationType>();

const defaultStylingForChar = new Map<string, string>([
    ['*', 'bold'],
    ['_', 'underlined'],
    ['#', 'inverse'],
]);

const lastUsedStylingsForEditor = new WeakMap<vscode.TextEditor, Set<string>>();

export function activateStyledown(context: vscode.ExtensionContext) {
    let activeEditor = vscode.window.activeTextEditor;

    if (activeEditor) {
        updateDecorations(activeEditor);
    }

    vscode.window.onDidChangeActiveTextEditor((editor) => {
        activeEditor = editor;
        if (activeEditor) {
            updateDecorations(activeEditor);
        }
    }, null, context.subscriptions);

    vscode.workspace.onDidChangeTextDocument((event) => {
        // If we ever need to work with very large Styledown documents and
        // performance becomes a concern, we can cache decorations and use
        // event.contentChanges to decide which ones need updating.
        if (activeEditor && event.document === activeEditor.document) {
            updateDecorations(activeEditor);
        }
    }, null, context.subscriptions);

    context.subscriptions.push(
        vscode.languages.registerFoldingRangeProvider(
            { language: 'styledown' }, { provideFoldingRanges }));

    context.subscriptions.push(
        vscode.languages.registerCodeLensProvider(
            { language: 'styledown' }, { provideCodeLenses }));

    return () => {
        for (const t of decorationTypeForStyling.values()) {
            t.dispose();
        }
        decorationTypeForStyling.clear();
    };
}

function updateDecorations(editor: vscode.TextEditor) {
    if (!editor || editor.document.languageId !== 'styledown') {
        return;
    }
    const document = editor.document;
    const contentLines = countContentLines(document);
    const stylingForChar = new Map(defaultStylingForChar);
    for (let i = contentLines; i < document.lineCount; i++) {
        const match = document.lineAt(i).text.match(/^(\S)\s+(.+)$/);
        if (match) {
            stylingForChar.set(match[1], match[2]);
        }
    }
    for (const styling of stylingForChar.values()) {
        if (!decorationTypeForStyling.has(styling)) {
            const t = vscode.window.createTextEditorDecorationType(applyStyling({}, styling));
            decorationTypeForStyling.set(styling, t);
        }
    }

    const rangesForStyling = new Map<string, vscode.Range[]>();
    for (let i = 0; i < contentLines; i += 2) {
        const styleLine = document.lineAt(i + 1).text;
        for (let j = 0; j < styleLine.length; j++) {
            const styling = stylingForChar.get(styleLine[j]);
            if (!styling) {
                continue;
            }
            let ranges = rangesForStyling.get(styling);
            if (!ranges) {
                rangesForStyling.set(styling, ranges = []);
            }
            ranges.push(new vscode.Range(i, j, i, j + 1));
        }
    }
    const lastUsedStylings = lastUsedStylingsForEditor.get(editor) || new Set();
    for (const [styling, ranges] of rangesForStyling.entries()) {
        editor.setDecorations(decorationTypeForStyling.get(styling)!, ranges);
        lastUsedStylings.delete(styling);
    }
    for (const styling of lastUsedStylings) {
        editor.setDecorations(decorationTypeForStyling.get(styling)!, []);
    }
    lastUsedStylingsForEditor.set(editor, new Set(rangesForStyling.keys()));
}

function provideFoldingRanges(document: vscode.TextDocument, context: vscode.FoldingContext, token: vscode.CancellationToken): vscode.FoldingRange[] {
    const ranges = [];
    for (let i = 0; i < countContentLines(document); i += 2) {
        ranges.push(new vscode.FoldingRange(i, i + 1));
    }
    return ranges;
}

function provideCodeLenses(document: vscode.TextDocument, token: vscode.CancellationToken): vscode.CodeLens[] {
    const range = new vscode.Range(0, 0, 0, 0);
    return [
        new vscode.CodeLens(range, { title: 'Fold All', command: 'editor.foldAll' }),
        new vscode.CodeLens(range, { title: 'Unfold All', command: 'editor.unfoldAll' })];
}

function countContentLines(document: vscode.TextDocument): number {
    for (let i = 0; i + 1 < document.lineCount; i += 2) {
        if (document.lineAt(i).text === '' && document.lineAt(i + 1).text !== '') {
            return i;
        }
    }
    return document.lineCount - document.lineCount % 2;
}
