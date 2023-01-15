const { LanguageClient } = require("vscode-languageclient/node");
import * as vscode from 'vscode';

let client;

function activate(context) {
    // Creates an Elvish output channel.
    let elvishOutputChannel = vscode.window.createOutputChannel("Elvish");
   
    client = new LanguageClient(
        "elvish",
        "Elvish Language Server",
        { command: "elvish", args: ["-lsp"] },
        { documentSelector: [{ scheme: "file", language: "elvish" }] }
    );
    client.start();

    const command = "elvish.eval";
    const commandHandler = () => {

        const window = vscode.window;
        const activeEditor = window.activeTextEditor;
        if (activeEditor && client) {
            const { text } = activeEditor.document.lineAt(activeEditor.selection.active.line);
            client.sendRequest("elvish/eval", {code: text})
                .then(result => {
                    const resultString = JSON.stringify(result);
                    elvishOutputChannel.appendLine(resultString);
                });
        }
    };
    context.subscriptions.push(vscode.commands.registerCommand(command, commandHandler));
}

function deactivate() {
    if (client) {
        return client.stop();
    }
}

module.exports = {
    activate,
    deactivate,
};
