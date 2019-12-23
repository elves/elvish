package web

const mainPageHTML = `<html>

  <body class="light">
    <div id="content">
      <div id="scrollback">
        <div id="progress"></div>
      </div>

      <div id="cmd" class="cmd">
        <span id="prompt" class="prompt">&gt;&gt;</span>
        <textarea id="buffer" class="buffer" rows="1"></textarea>
      </div>

      <div id="theme-switchers" class="flex">
        <span class="theme-switcher" id="dark-theme">dark</span>
        <span class="theme-switcher" id="light-theme">light</span>
      </div>

    </div>
  </body>

  <style>
    /* Colors are taken from Material palette. */

    /* Global styles */

    * {
      margin: 0;
      padding: 0;
      font: 12pt monospace;
    }

    #content {
      margin: 20px;
      padding: 20px;
    }

    /* Scrollback */

    .exception {
      font-weight: bold;
    }

    .server-error {
      font-style: italic;
    }

    /* Command line */

    .cmd {
      display: flex;
      width: 100%;
    }

    .cmd > .prompt {
      display: inline-block;
      margin-right: 1em;
    }

    .cmd > .buffer {
      flex: 1;
    }

    /* Theme switcher */

    #theme-switchers {
      display: flex;
      margin-top: 0.4em;
    }

    .theme-switcher {
      cursor: pointer;
      padding: 0 1em;
      border: 1px solid;
    }

    #light-theme {
      color: black;
      background-color: white;
      border-color: black;
    }

    #dark-theme {
      color: white;
      background-color: black;
      border-color: white;
    }

    /* Color schemes. Color values from Material palette. */

    body.light {
      background: #EEEEEE; /* grey 200 */
    }

    .light * {
      color: black;
      background: white;
    }

    .light #content {
      background: white;
    }

    .light .cmd > .prompt {
      color: #1565C0; /* blue 800 */
    }

    .light .cmd > #prompt {
      color: #2E7D32; /* green 800 */
    }

    .light .error {
      color: #C62828; /* red 800 */
    }


    .dark * {
      color: white;
      background: black;
    }

    body.dark {
      background: #212121; /* grey 900 */
    }

    .dark #content {
      background: black;
    }

    .dark .cmd > .prompt {
      color: #90CAF9; /* blue 200 */
    }

    .dark .cmd > #prompt {
      color: #A5D6A7; /* green 200 */
    }

    .dark .error {
      color: #EF9A9A; /* red 200 */
    }

  </style>

  <script>
    // TODO(xiaq): Stream results.
    var $prompt = document.getElementById('prompt'),
        $buffer = document.getElementById('buffer'),
        $scrollback = document.getElementById('scrollback'),
        $progress = document.getElementById('progress');

    /* Theme switchers. */
    document.getElementById('dark-theme').onclick = function() {
      document.body.className = 'dark';
    };

    document.getElementById('light-theme').onclick = function() {
      document.body.className = 'light';
    };

    $buffer.addEventListener('keypress', function(e) {
      if (e.keyCode == 13 &&
          !e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey) {
        e.preventDefault();
        execute();
      }
    });

    function execute() {
      var code = $buffer.value;
      addToScrollbackInner(freezeCmd());
      $buffer.value = '';
      $progress.innerText = 'executing...';

      var req = new XMLHttpRequest();
      req.onloadend = function() {
        $progress.innerText = '';
      };
      req.onload = function() {
        var res = JSON.parse(req.responseText);
        addToScrollback('output', res.OutBytes);
        if (res.OutValues) {
          for (var v of res.OutValues) {
            addToScrollback('output-value', v);
          }
        }
        addToScrollback('error', res.ErrBytes);
        addToScrollback('error exception', res.Err);
      };
      req.onerror = function() {
        addToScrollback('error server-error', req.responseText
          || req.statusText
          || (req.status == req.UNSENT && "lost connection")
          || "unknown error");
      };
      req.open('POST', '/execute');
      req.send(code);
    }

    function addToScrollback(className, innerText) {
      var div = document.createElement('div');
      div.className = className;
      div.innerText = innerText;

      addToScrollbackInner(div);
    }

    function addToScrollbackInner(element) {
      $scrollback.insertBefore(element, $progress);
      window.scrollTo(0, document.body.scrollHeight);
    }

    function freezeCmd() {
      var cmd = document.createElement('div'),
          prompt = document.createElement('span'),
          buffer = document.createElement('span');
      cmd.className = 'cmd';
      prompt.className = 'prompt';
      prompt.innerText = $prompt.innerText;
      buffer.className = 'buffer';
      buffer.innerText = $buffer.value;
      cmd.appendChild(prompt);
      cmd.appendChild(buffer);
      return cmd;
    }

  </script>

</html>
`

// vim: se ft=html si et sw=2 ts=2 sts=2:
