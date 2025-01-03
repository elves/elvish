/** Entrypoint of the extension. */

import * as vscode from 'vscode';

import { activateStyledown } from './styledown';
import { activateLogging } from './utils/logging';
import { activateLsp } from './lsp';
import { activateTranscript } from './transcript';

type Activator = (context: vscode.ExtensionContext) => Deactivator;
type Deactivator = (() => void);

const activators: Activator[] = [
    activateLogging, // This must come first as other activators might depend on it
    activateLsp, activateStyledown, activateTranscript
];

const deactivators: Deactivator[] = [];

export function activate(context: vscode.ExtensionContext) {
    for (const activate of activators) {
        const deactivate = activate(context);
        deactivators.unshift(deactivate);
    }
}

export function deactivate() {
    for (const deactivate of deactivators) {
        deactivate();
    }
}
