const { LanguageClient } = require("vscode-languageclient/node");

let client;

function activate() {
    client = new LanguageClient(
        "elvish",
        "Elvish Language Server",
        { command: "elvish", args: ["-lsp"] },
        { documentSelector: [{ scheme: "file", language: "elvish" }] }
    );
    client.start();
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
