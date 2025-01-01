import * as vscode from 'vscode';

export let debugChannel: vscode.OutputChannel;

export function activateLogging(context: vscode.ExtensionContext) {
    debugChannel = vscode.window.createOutputChannel('Elvish Debug');
    return () => { };
}
