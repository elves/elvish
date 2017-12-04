package web

const mainPageHTML = `<html>

  <div id="scrollback">
    <div id="progress"></div>
  </div>
  <textarea id="code" rows="4"></textarea>

  <style>
    #code {
      width: 100%;
    }
    #scrollback, #code, #progress {
      font-family: monospace;
      font-size: 11pt;
    }
    .code {
      color: white;
      background-color: black;
    }
    .error, .exception {
      color: red;
    }
    .exception {
      font-weight: bold;
    }
    .server-error {
      color: white;
      background-color: red;
    }
  </style>

  <script>
    // TODO(xiaq): Stream results.
    var $scrollback = document.getElementById('scrollback'),
        $code = document.getElementById('code'),
        $progress = document.getElementById('progress');

    $code.addEventListener('keypress', function(e) {
      if (e.keyCode == 13 &&
          !e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey) {
        e.preventDefault();
        execute();
      }
    });

    function execute() {
      var code = $code.value;
      $code.value = '';
      addToScrollback('code', code);
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
        addToScrollback('exception', res.Err);
      };
      req.onerror = function() {
        addToScrollback('server-error', req.responseText
          || req.statusText
          || (req.status == req.UNSENT && "lost connection")
          || "unknown error");
      };
      req.open('POST', '/execute');
      req.send(code);
    }

    function addToScrollback(className, innerText) {
      var $div = document.createElement('div')
      $div.className = className;
      $div.innerText = innerText;
      $scrollback.insertBefore($div, $progress);

      window.scrollTo(0, document.body.scrollHeight);
    }

  </script>

</html>
`
