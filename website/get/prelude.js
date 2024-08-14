document.addEventListener('DOMContentLoaded', main);

const binaryAvailable = new Set([
  'linux-amd64', 'linux-386', 'linux-arm64', 'linux-riscv64',
  'darwin-amd64', 'darwin-arm64',
  'freebsd-amd64',
  'netbsd-amd64',
  'openbsd-amd64',
  'windows-amd64', 'windows-386',
]);

function main() {
  // Set up change detection.
  for (const e of document.querySelectorAll('input')) {
    e.addEventListener('input', (event) => {
      const el = event.target;
      onChange(el.name, el.value);
    });
  }
  // Populate os and arch, either from localStorage or by auto-detection.
  const os = tryGetLocalStorage('os');
  const arch = tryGetLocalStorage('arch');
  if (os && arch && binaryAvailable.has(os + '-' + arch)) {
    select('os', os);
    select('arch', arch);
  } else {
    autoDetectOsAndArch();
  }
  // Populate version and dir from localStorage.
  const version = tryGetLocalStorage('version');
  if (version) {
    select('version', version);
  }
  // Populate dir, sudo and mirror from localStorage, and open the <details> if
  // they have non-default values.
  var openDetails = false;
  const dir = tryGetLocalStorage('dir');
  if (dir) {
    document.querySelector('input[name="dir"]').value = dir;
    onChange('dir', dir);
    openDetails = true;
  }
  const sudo = tryGetLocalStorage('sudo');
  if (sudo && sudo !== 'sudo') {
    select('sudo', sudo);
    openDetails = true;
  }
  const mirror = tryGetLocalStorage('mirror');
  if (mirror && mirror !== 'official') {
    select('mirror', mirror);
    openDetails = true;
  }
  if (openDetails) {
    document.querySelector('details').open = true;
  }
}

function onChange(name, value) {
  trySetLocalStorage(name, value);

  // Update input controls.
  if (name === 'os') {
    const os = value;
    // Disable unsupported architectures.
    for (const element of document.querySelectorAll('input[name="arch"]')) {
      element.disabled = !binaryAvailable.has(os + '-' + element.value);
      if (element.disabled && element.checked) {
        const fallbackArch = fallbackArchForOs(os);
        select('arch', fallbackArch, {suppressOnChange: true});
        // Because we suppressed the recursive call to onChange, we have to
        // store the new arch manually.
        trySetLocalStorage('arch', fallbackArch);
      }
    }
    // Update directory placeholder and the installation instruction.
    const $dir = document.querySelector('input[name="dir"]');
    const $where = document.querySelector('#where');
    if (os === 'windows') {
      $dir.placeholder = '$Env:USERPROFILE\\Utilities';
      $where.innerText = 'PowerShell';
    } else {
      $dir.placeholder = '/usr/local/bin';
      $where.innerText = 'a terminal';
    }
  }

  // Update outputs.
  const $form = document.querySelector('form');
  const f = new FormData($form);

  const $script = document.querySelector('#script');
  $script.innerHTML = genScriptHTML(
    f.get('os'), f.get('arch'), f.get('version'),
    f.get('dir') || document.querySelector('input[name="dir"]').placeholder,
    f.get('sudo'), f.get('mirror'));
}

function genScriptHTML(os, arch, version, dir, sudo, mirror) {
  const host = mirror === 'tuna' ? 'mirrors.tuna.tsinghua.edu.cn/elvish' : 'dl.elv.sh';
  const urlBase = `https://${host}/${os}-${arch}/elvish-${version}`;
  if (os === 'windows') {
    const url = link(urlBase + '.zip');
    return `& {
md "${dir}" -force > $null
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if (!(($UserPath -split ';') -contains "${dir}")) {
  [Environment]::SetEnvironmentVariable("PATH", $UserPath + ";${dir}", "User")
  $Env:PATH += ";${dir}"
}
cd "${dir}"
Invoke-RestMethod -Uri '${url}' -OutFile elvish.zip
Expand-Archive -Force elvish.zip -DestinationPath .
rm elvish.zip
}`;
  } else {
    const url = link(urlBase + '.tar.gz');
    const sudoPrefix = sudo == 'dont' ? '' : sudo + ' ';
    return highlightSh(`curl -so - ${url} | ${sudoPrefix}tar -xzvC ${dir}`);
  }
}

function link(s) {
  return `<a href="${s}">${s}</a>`;
}

function highlightSh(s) {
  // Use a simplistic algorithm to color command names and pipes green.
  return s.replace(/(^|\|\s*)[\w_-]+/mg, '<span class="sgr-32">$&</span>')
}

function autoDetectOsAndArch() {
  const os = detectOs(navigator.platform);
  if (os) {
    select('os', os);
    // Select fallback and trigger change detection first in case the promise
    // errors or never resolves.
    select('arch', fallbackArchForOs(os));

    let dataPromise = Promise.resolve();
    if (navigator.userAgentData && navigator.userAgentData.getHighEntropyValues) {
      dataPromise = navigator
        .userAgentData.getHighEntropyValues(['architecture', 'bitness']);
    }

    dataPromise.then((data) => {
      const arch = detectArch(navigator.platform, data) || fallbackArchForOs(os);
      if (binaryAvailable[os+'-'+arch]) {
        select('arch', arch);
      }
    }).catch(() => {
      // Do nothing
    });
  } else {
    // Use fallback and trigger change detection.
    select('os', 'linux');
    select('arch', fallbackArchForOs('linux'));
  }
}

// Detects GOOS from navigator.platform. Partially based on
// https://stackoverflow.com/a/19883965/566659.
//
// There is a better defined value in navigator.userAgentData.platform, but only
// Chrome supports it.
function detectOs(p) {
  if (p.match(/^linux/i)) {
    return 'linux';
  } else if (p.match(/^mac/i)) {
    return 'darwin';
  } else if (p.match(/^freebsd/i)) {
    return 'freebsd';
  } else if (p.match(/^netbsd/i)) {
    // Not in the StackOverflow answer, but a reasonable guess
    return 'netbsd';
  } else if (p.match(/^openbsd/i)) {
    return 'openbsd';
  } else if (p.match(/^win/i)) {
    return 'windows';
  }
}

// Detects GOARCH from navigator.platform and the high entropy data (if
// available).
//
// TODO: Add detection code for RISC-V.
function detectArch(p, data) {
  if (data) {
    const arch = {
      arm_64: 'arm64',
      x86_64: 'amd64',
      x86_32: '386'
    }[data.architecture + '_' + data.bitness];
    if (arch) {
      return arch;
    }
  }
  if (p.match(/aarch64/i)) {
    return 'arm64';
  } else if (p.match(/x86_64|amd64/i)) {
    return 'amd64';
  } else if (p.match(/i[3456]86/i)) {
    return '386';
  }
}

function fallbackArchForOs(os) {
  if (os === 'darwin') {
    return 'arm64';
  } else {
    return 'amd64';
  }
}

function select(name, value, opts) {
  const element = document.querySelector(
    `input[name="${name}"][value="${value}"]`)
  if (element) {
    element.checked = true;
    if (opts && opts.suppressOnChange) {
      // Don't call onChange
    } else {
      onChange(name, value);
    }
  }
}

// localStorage may not be supported by the browser, or its access may be denied
// due to security settings. Wrap it to swallow exceptions.

function trySetLocalStorage(key, value) {
  try {
    localStorage.setItem(key, value);
  } catch (e) {
  }
}

function tryGetLocalStorage(key) {
  try {
    return localStorage.getItem(key)
  } catch (e) {
  }
}

function copyScript(event) {
  event.preventDefault();

  // Based on https://stackoverflow.com/a/48020189/566659
  window.getSelection().removeAllRanges(); // clear current selection
  const range = document.createRange();
  range.selectNode(document.getElementById("script"));
  window.getSelection().addRange(range); // select text
  document.execCommand("copy");
  window.getSelection().removeAllRanges();// clear selection again

  const oldText = event.target.innerText;
  event.target.innerText = 'copied!';
  setTimeout(() => {
    event.target.innerText = oldText;
  }, 1500);
}
