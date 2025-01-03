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
            { language: 'styledown' }, styledownProvider));
    context.subscriptions.push(
        vscode.languages.registerCodeLensProvider(
            { language: 'styledown' }, styledownProvider));

    context.subscriptions.push(
        vscode.languages.registerFoldingRangeProvider(
            { language: 'elvish-transcript' }, elvishTranscriptProvider));
    context.subscriptions.push(
        vscode.languages.registerCodeLensProvider(
            { language: 'elvish-transcript' }, elvishTranscriptProvider));


    return () => {
        for (const t of compiledStyling.values()) {
            t.dispose();
        }
        compiledStyling.clear();
    };
}

// The default styling characters implemented by the styledown package.
const defaultStylingForChar = new Map<string, string>([
    ['*', 'bold'],
    ['_', 'underlined'],
    ['#', 'inverse'],
]);

// A global cache of styling strings "compiled" to TextEditorDecorationType.
//
// This can accmulate unused stylings as the user edits the configuration
// stanza. If this becomes a concern, we can add reference counting for each
// styling and dispose of those that are no longer referenced.
//
// This needs to be declared here because the variable below this one depends on
// it being initialized first.
const compiledStyling = new Map<string, vscode.TextEditorDecorationType>();

function compileStyling(s: string): vscode.TextEditorDecorationType {
    return getOrSet(compiledStyling, s, () =>
        vscode.window.createTextEditorDecorationType(applyStyling({}, s)));
}

// The extended set of characters used by the "render" command in etktest.go.
// This never changes, so we can precompile it.
const etktestDecorationTypeForChar = mapValues([
    ['*', 'bold'],
    ['_', 'underlined'],
    ['#', 'inverse'],
    ['R', 'red'],
    ['G', 'green'],
    ['M', 'magenta'],
], compileStyling);

function updateDecorations(editor: vscode.TextEditor) {
    const id = editor.document.languageId;
    if (id === 'styledown') {
        updateDecorationsStyledown(editor);
    } else if (id === 'elvish-transcript') {
        updateDecorationsElvishTranscript(editor);
    }
}

function updateDecorationsStyledown(editor: vscode.TextEditor) {
    const document = editor.document;
    const contentLines = countContentLines(document);

    // 'R' -> 'bold fg-red'
    const stylingForChar = new Map(defaultStylingForChar);
    for (let i = contentLines; i < document.lineCount; i++) {
        const match = document.lineAt(i).text.match(/^(\S)\s+(.+)$/);
        if (match) {
            stylingForChar.set(match[1], match[2]);
        }
    }

    // 'R' -> corresponding TextEditorDecorationType t_R
    const decorationTypeForChar = mapValues(stylingForChar, compileStyling);

    // t_R -> [new Range(0, 0, 0, 1), ...]
    const rangesForDecorationType = new Map<vscode.TextEditorDecorationType, vscode.Range[]>();
    for (let i = 0; i < contentLines; i += 2) {
        const styleLine = document.lineAt(i + 1).text;
        for (let j = 0; j < styleLine.length; j++) {
            const t = decorationTypeForChar.get(styleLine[j]);
            if (!t) {
                continue;
            }
            getOrSet(rangesForDecorationType, t, () => [])
                .push(new vscode.Range(i, j, i, j + 1));
        }
    }
    setDecorations(editor, rangesForDecorationType);
}

const cursorPlaceholder = '̅̂';

const cursorDecorationType = vscode.window.createTextEditorDecorationType({
    borderColor: new vscode.ThemeColor('editor.foreground'),
    borderWidth: '1px',
    borderStyle: 'solid',
});

function updateDecorationsElvishTranscript(editor: vscode.TextEditor) {
    const document = editor.document;

    const rangesForDecorationType = new Map<vscode.TextEditorDecorationType, vscode.Range[]>();
    for (const { start, end } of findTtyBlocks(document)) {
        for (let i = start; i < end; i += 2) {
            let styleLine = document.lineAt(i + 1).text;
            const cursorIdx = styleLine.indexOf(cursorPlaceholder);
            if (cursorIdx != -1) {
                getOrSet(rangesForDecorationType, cursorDecorationType, () => [])
                    .push(new vscode.Range(i, cursorIdx - 1, i, cursorIdx));
                styleLine = styleLine.replace(cursorPlaceholder, '');
            }
            for (let j = 1; j < styleLine.length - 1; j++) {
                const t = etktestDecorationTypeForChar.get(styleLine[j]);
                if (!t) {
                    continue;
                }
                getOrSet(rangesForDecorationType, t, () => [])
                    .push(new vscode.Range(i, j, i, j + 1));
            }
        }
    }
    setDecorations(editor, rangesForDecorationType);
}

const lastUsedDecorationTypesForEditor = new WeakMap<vscode.TextEditor, Set<vscode.TextEditorDecorationType>>();

// VS Code's TextEditor retains all the ranges associated with a
// TextEditorDecorationType; calling editor.setDecorations(t, ...) will
// implicitly clear all previous ranges associated with t. But there is no easy
// way of clearing decorations with a certain t that is no longer used.
//
// This function takes care of this by remembering the set of
// TextEditorDecorationType's passed to the last call for the same editor.
function setDecorations(editor: vscode.TextEditor, rangesForDecorationType: Map<vscode.TextEditorDecorationType, vscode.Range[]>) {
    const lastUsedDecorationTypes = lastUsedDecorationTypesForEditor.get(editor) || new Set();
    for (const [t, ranges] of rangesForDecorationType.entries()) {
        editor.setDecorations(t, ranges);
        lastUsedDecorationTypes.delete(t);
    }
    for (const t of lastUsedDecorationTypes) {
        editor.setDecorations(t, []);
    }
    lastUsedDecorationTypesForEditor.set(editor, new Set(rangesForDecorationType.keys()));
}

const styledownProvider = {
    provideFoldingRanges(document: vscode.TextDocument, context: vscode.FoldingContext, token: vscode.CancellationToken): vscode.FoldingRange[] {
        const ranges = [];
        for (let i = 0; i < countContentLines(document); i += 2) {
            ranges.push(new vscode.FoldingRange(i, i + 1));
        }
        return ranges;
    },
    provideCodeLenses(document: vscode.TextDocument, token: vscode.CancellationToken): vscode.CodeLens[] {
        const range = new vscode.Range(0, 0, 0, 0);
        return [
            new vscode.CodeLens(range, { title: 'Fold style lines', command: 'editor.foldAll' }),
            new vscode.CodeLens(range, { title: 'Unfold style lines', command: 'editor.unfoldAll' })];
    }
}

const elvishTranscriptProvider = {
    provideFoldingRanges(document: vscode.TextDocument, context: vscode.FoldingContext, token: vscode.CancellationToken): vscode.FoldingRange[] {
        const ranges = [];
        for (const { start, end } of findTtyBlocks(document)) {
            for (let i = start; i < end; i += 2) {
                ranges.push(new vscode.FoldingRange(i, i + 1));
            }
        }
        return ranges;
    },
    provideCodeLenses(document: vscode.TextDocument, token: vscode.CancellationToken): vscode.CodeLens[] {
        const lenses = [];
        const allSelectionLines = [];
        for (const { start, end } of findTtyBlocks(document)) {
            const range = new vscode.Range(start - 1, 0, start - 1, 0);
            const selectionLines = [];
            for (let i = start; i < end; i += 2) {
                selectionLines.push(i);
                allSelectionLines.push(i);
            }
            lenses.push(
                new vscode.CodeLens(range, {
                    title: 'Fold style lines',
                    command: 'editor.fold', arguments: [{ selectionLines }]
                }),
                new vscode.CodeLens(range, {
                    title: 'Unfold style lines',
                    command: 'editor.unfold', arguments: [{ selectionLines }]
                }));
        }

        const range = new vscode.Range(0, 0, 0, 0);
        lenses.push(
            new vscode.CodeLens(range, {
                title: 'Fold all style lines',
                command: 'editor.fold', arguments: [{ selectionLines: allSelectionLines }]
            }),
            new vscode.CodeLens(range, {
                title: 'Unfold all style lines',
                command: 'editor.unfold', arguments: [{ selectionLines: allSelectionLines }]
            }));
        return lenses;
    }
}


function countContentLines(document: vscode.TextDocument): number {
    for (let i = 0; i + 1 < document.lineCount; i += 2) {
        if (document.lineAt(i).text === '' && document.lineAt(i + 1).text !== '') {
            return i;
        }
    }
    return document.lineCount - document.lineCount % 2;
}

function findTtyBlocks(document: vscode.TextDocument): { start: number, end: number }[] {
    const blocks = [];
    for (let i = 0; i < document.lineCount; i++) {
        if (document.lineAt(i).text.startsWith('┌')) {
            let j;
            for (j = i + 1; j < document.lineCount; j++) {
                const line = document.lineAt(j).text;
                if (line.startsWith('└')) {
                    blocks.push({ start: i + 1, end: j });
                    break;
                } else if (!line.startsWith('│')) {
                    break;
                }
            }
            i = j;
        }
    }
    return blocks;
}

function mapValues<K, V, U>(m: Iterable<[K, V]>, f: (v: V) => U): Map<K, U> {
    return new Map(Array.from(m, ([k, v]) => [k, f(v)]));
}

function getOrSet<K, V>(m: Map<K, V>, k: K, f: () => V): V {
    const v = m.get(k);
    if (v) {
        return v;
    }
    const newv = f();
    m.set(k, newv);
    return newv;
}
