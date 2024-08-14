<div class="has-js">
<div class="form-wrapper">
<form>

  <div class="control">
    <header>OS</header>
    <div class="widgets">
      <label class="option">
        <input type="radio" name="os" value="linux"/>
        Linux
      </label>
      <label class="option">
        <input type="radio" name="os" value="darwin"/>
        macOS
      </label>
      <label class="option">
        <input type="radio" name="os" value="freebsd"/>
        FreeBSD
      </label>
      <label class="option">
        <input type="radio" name="os" value="netbsd"/>
        NetBSD
      </label>
      <label class="option">
        <input type="radio" name="os" value="openbsd"/>
        OpenBSD
      </label>
      <label class="option">
        <input type="radio" name="os" value="windows"/>
        Windows
      </label>
    </div>
  </div>

  <div class="control">
    <header>CPU</header>
    <div class="widgets">
      <label class="option">
        <input type="radio" name="arch" value="amd64"/>
        Intel 64-bit
      </label>
      <label class="option">
        <input type="radio" name="arch" value="arm64"/>
        ARM 64-bit
      </label>
      <label class="option">
        <input type="radio" name="arch" value="386"/>
        Intel 32-bit
      </label>
      <label class="option">
        <input type="radio" name="arch" value="riscv64"/>
        RISC-V
      </label>
    </div>
  </div>

  <div class="small-print">

If your OS/CPU combination is missing or grayed out, you may still be able to
<a href="https://github.com/elves/elvish/blob/master/docs/building.md" target="_blank">build
Elvish from source</a>.

  </div>

  <div class="control">
    <header>Version</header>
    <div class="widgets">
      <label class="option">
        <input type="radio" name="version" value="v0.21.0" checked/>
        0.21.0
      </label>
      <label class="option">
        <input type="radio" name="version" value="HEAD"/>
        HEAD
      </label>
    </div>
  </div>

  <div class="small-print">

0.21.0 is the [latest release](../blog/0.21.0-release-notes.html). Suitable if
you prefer to only update occasionally or need a stable scripting environment.

HEAD is the latest development build. It has the
<a href="https://github.com/elves/elvish/blob/master/0.22.0-release-notes.md" target="_blank">freshest
features</a> and is stable enough for interactive use.

  </div>

  <details>
    <summary>Customize the script</summary>
    <div class="advanced">
      <div class="control">
        <header>Install to</header>
        <div class="widgets">
          <input type="text" name="dir" placeholder="/usr/local/bin" />        
        </div>
      </div>
      <div class="control">
        <header>Sudo</header>
        <div class="widgets">
          <label class="option">
            <input type="radio" name="sudo" value="sudo" checked/>
            use <code>sudo</code>
          </label>
          <label class="option">
            <input type="radio" name="sudo" value="doas"/>
            use <code>doas</code>
          </label>
          <label class="option">
            <input type="radio" name="sudo" value="dont"/>
            don't use
          </label>
        </div>
      </div>
      <div class="small-print">
        Choose “don’t use” if you are running as
        <code>root</code> or installing to a directory you can write to.
        No effect on Windows.
      </div>
      <div class="control">
        <header>Mirror</header>
        <div class="widgets">
          <label class="option">
            <input type="radio" name="mirror" value="official" checked/>
            official
          </label>
          <label class="option">
            <input type="radio" name="mirror" value="tuna"/>
            TUNA
          </label>
        </div>
      </div>
      <div class="small-print">
        The <a href="https://mirrors.tuna.tsinghua.edu.cn" target="_blank">TUNA mirror site</a>
        is hosted in Tsinghua University, Beijing, China.
      </div>
    </div>
  </details>

</form>
</div>
<div class="content">

Run the following in <span id="where">a terminal</span> to install Elvish
(<a href="#" onclick="copyScript(event)">copy to clipboard</a>):

<pre><code id="script">
</code></pre>

Alternative, click the link above to download the archive and unpack it in
directory in `PATH`.

More topics about installing Elvish:

</div>
</div>
<div class="no-js">
<div class="content">

Enable JavaScript to generate an installation script for your platform.

Alternatively, find your binary for your platform in the
[all binaries](all-binaries.html) page and unpack it manually.

</div>
</div>
