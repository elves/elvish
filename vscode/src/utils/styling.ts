/** Conceptually, a TypeScript reimplementation of pkg/ui/styling.go. */

import * as vscode from 'vscode';

// apply(style, stylings) is like ApplyStyling(style, ParseStyling(stylings)) in
// Go.
// 
// Currently only a small subset is implemented.
export function applyStyling(style: vscode.DecorationRenderOptions, stylings: string): vscode.DecorationRenderOptions {
    style = Object.assign({}, style);
    for (const styling of stylings.split(/\s+/)) {
        if (styling === 'bold') {
            style.fontWeight = 'bold';
        } else if (styling === 'underlined') {
            style.textDecoration = 'underline';
        } else if (styling === 'inverse') {
            [style.color, style.backgroundColor] = [
                style.backgroundColor || new vscode.ThemeColor('editor.background'),
                style.color || new vscode.ThemeColor('editor.foreground')];
        } else if (styling.startsWith('bg-')) {
            style.backgroundColor = styling.substring(3);
        } else if (styling.startsWith('fg-')) {
            style.color = styling.substring(3);
        } else {
            style.color = styling;
        }
    }
    return style;
}
