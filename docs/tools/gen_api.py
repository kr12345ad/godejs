#!/usr/bin/env python3
"""Generate the complete API reference from a runtime snapshot and Node 24 docs.

Usage, from the repository root:

    ./gode docs/tools/dump-api.js > docs/tools/api.json
    git clone --depth 1 --filter=blob:none --sparse \
        --branch v24.22.2 https://github.com/nodejs/node /tmp/node-docs
    git -C /tmp/node-docs sparse-checkout set doc/api
    python3 docs/tools/gen_api.py

The runtime snapshot records every own property of every builtin and every host
global. Node's own API documentation supplies signatures, parameter types,
return descriptions, and examples. Entries with no upstream section (private
helpers and GodeJS extensions) are marked instead of being given invented
behaviour. Generated pages are checked in, so a normal Sphinx build needs
neither a compiled binary nor the Node source checkout.
"""

import html
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
API_JSON = os.path.join(ROOT, "docs", "tools", "api.json")
OUT = os.path.join(ROOT, "docs", "usage", "api")
NODE_DOCS = os.environ.get("NODE_DOCS", "/tmp/node-docs/doc/api")

FOLDED = {"assert/strict", "fs/promises", "path/posix", "path/win32", "sys"}

TITLES = {
    "assert": "assert",
    "async_hooks": "async_hooks",
    "buffer": "buffer",
    "child_process": "child_process",
    "cluster": "cluster",
    "console": "console",
    "constants": "constants",
    "crypto": "crypto",
    "dgram": "dgram",
    "diagnostics_channel": "diagnostics_channel",
    "dns": "dns",
    "domain": "domain",
    "events": "events",
    "fs": "fs",
    "http": "http",
    "http2": "http2",
    "http3": "http3",
    "https": "https",
    "inspector": "inspector",
    "module": "module",
    "net": "net",
    "os": "os",
    "path": "path",
    "perf_hooks": "perf_hooks",
    "process": "process",
    "punycode": "punycode",
    "querystring": "querystring",
    "readline": "readline",
    "repl": "repl",
    "sea": "sea",
    "sqlite": "sqlite",
    "stream": "stream",
    "string_decoder": "string_decoder",
    "test": "test (node:test)",
    "timers": "timers",
    "tls": "tls",
    "trace_events": "trace_events",
    "tty": "tty",
    "url": "url",
    "util": "util",
    "v8": "v8",
    "vm": "vm",
    "wasi": "wasi",
    "worker_threads": "worker_threads",
    "zlib": "zlib",
}

DOC_FILES = {
    "assert": "assert.md",
    "async_hooks": "async_hooks.md",
    "buffer": "buffer.md",
    "child_process": "child_process.md",
    "cluster": "cluster.md",
    "console": "console.md",
    "crypto": "crypto.md",
    "dgram": "dgram.md",
    "diagnostics_channel": "diagnostics_channel.md",
    "dns": "dns.md",
    "domain": "domain.md",
    "events": "events.md",
    "fs": "fs.md",
    "http": "http.md",
    "http2": "http2.md",
    "https": "https.md",
    "inspector": "inspector.md",
    "module": "module.md",
    "net": "net.md",
    "os": "os.md",
    "path": "path.md",
    "perf_hooks": "perf_hooks.md",
    "process": "process.md",
    "punycode": "punycode.md",
    "querystring": "querystring.md",
    "readline": "readline.md",
    "repl": "repl.md",
    "sea": "single-executable-applications.md",
    "sqlite": "sqlite.md",
    "stream": "stream.md",
    "string_decoder": "string_decoder.md",
    "test": "test.md",
    "timers": "timers.md",
    "tls": "tls.md",
    "tty": "tty.md",
    "url": "url.md",
    "util": "util.md",
    "v8": "v8.md",
    "vm": "vm.md",
    "wasi": "wasi.md",
    "worker_threads": "worker_threads.md",
    "zlib": "zlib.md",
}

GROUPS = [
    ("Core runtime", [
        "assert", "buffer", "console", "events", "fs", "path", "process",
        "querystring", "stream", "string_decoder", "timers", "tty", "url",
        "util",
    ]),
    ("Networking", [
        "dgram", "dns", "http", "http2", "http3", "https", "net", "tls",
    ]),
    ("Concurrency and loading", [
        "async_hooks", "child_process", "cluster", "module", "os", "readline",
        "repl", "sqlite", "test", "vm", "worker_threads",
    ]),
    ("Platform and diagnostics", [
        "constants", "crypto", "diagnostics_channel", "domain", "inspector",
        "perf_hooks", "punycode", "sea", "trace_events", "v8", "wasi", "zlib",
    ]),
]

KIND_LABEL = {
    "function": "function",
    "object": "object",
    "array": "array",
    "string": "string",
    "number": "number",
    "boolean": "boolean",
    "null": "null",
    "undefined": "undefined",
}

# GodeJS extensions and APIs that Node documents only by reference.
CURATED = {
    "http3": {
        "description": (
            "A native HTTP/3 server implemented with Go's ``quic-go`` package. "
            "HTTP/3 is not part of Node's standard library; this module exposes "
            "the same ``IncomingMessage`` and ``ServerResponse`` objects as "
            "``node:http``."
        ),
        "example": """const fs = require('node:fs');
const http3 = require('gode:http3');

const server = http3.createServer({
  key: fs.readFileSync('key.pem'),
  cert: fs.readFileSync('cert.pem'),
}, (req, res) => {
  res.setHeader('content-type', 'application/json');
  res.end(JSON.stringify({ ok: true, url: req.url }));
});

server.listen(4433, '0.0.0.0', () => {
  console.log('HTTP/3 on', server.address());
});""",
    },
    "http3.createServer": {
        "sig": "http3.createServer([options][, requestListener])",
        "description": "Create an HTTP/3 server. TLS options must include ``key`` and ``cert``; ``requestListener`` is called with ``(req, res)`` exactly like ``http.createServer``.",
        "args": [("options", "Object", True, "TLS key/certificate and HTTP/3 options."), ("requestListener", "Function", True, "Handler receiving IncomingMessage and ServerResponse.")],
        "returns": "http3.Server",
    },
    "http3.createSecureServer": {
        "sig": "http3.createSecureServer([options][, requestListener])",
        "description": "Alias of ``http3.createServer``. The listener is always TLS because QUIC requires encryption.",
        "args": [("options", "Object", True, "TLS key/certificate and HTTP/3 options."), ("requestListener", "Function", True, "Handler receiving IncomingMessage and ServerResponse.")],
        "returns": "http3.Server",
    },
    "http3.Server": {
        "sig": "new http3.Server([options][, requestListener])",
        "description": "The HTTP/3 server class. Its request, response, and lifecycle methods mirror ``http.Server``.",
        "args": [("options", "Object", True, "TLS and HTTP/3 options."), ("requestListener", "Function", True, "Request handler.")],
        "returns": "http3.Server",
    },
    "globalThis.localStorage": {
        "description": "A persistent Web Storage area. Values are strings, keys keep insertion order, and data is stored in the file selected by ``--localstorage-file`` or the per-working-directory default. ``localStorage`` is an object, not a function.",
    },
    "globalThis.sessionStorage": {
        "description": "An in-memory Web Storage area for the lifetime of the process. ``sessionStorage`` is an object, not a function.",
    },
    "globalThis.Storage": {
        "description": "The Web Storage interface implemented by ``localStorage`` and ``sessionStorage``. Constructing it directly is illegal; use one of the two global areas.",
    },
    "globalThis.localStorage.getItem": {
        "sig": "localStorage.getItem(key)",
        "description": "Return the value stored under ``key``, or ``null`` when the key does not exist.",
        "args": [("key", "string", False, "Key to look up; converted to a string.")],
        "returns": "string | null",
    },
    "globalThis.localStorage.setItem": {
        "sig": "localStorage.setItem(key, value)",
        "description": "Store ``value`` under ``key``. Both arguments are converted to strings.",
        "args": [("key", "string", False, "Key to write."), ("value", "string", False, "Value to store.")],
        "returns": "undefined",
    },
    "globalThis.localStorage.removeItem": {
        "sig": "localStorage.removeItem(key)",
        "description": "Delete ``key`` from the area. Removing a missing key is a no-op.",
        "args": [("key", "string", False, "Key to delete.")],
        "returns": "undefined",
    },
    "globalThis.localStorage.clear": {
        "sig": "localStorage.clear()",
        "description": "Delete every key in the area.",
        "returns": "undefined",
    },
    "globalThis.localStorage.key": {
        "sig": "localStorage.key(index)",
        "description": "Return the key at ``index`` in insertion order, or ``null`` when the index is out of range.",
        "args": [("index", "number", False, "Zero-based key index.")],
        "returns": "string | null",
    },
    "globalThis.localStorage.length": {
        "description": "The number of stored keys.",
    },
    "globalThis.sessionStorage.getItem": {
        "sig": "sessionStorage.getItem(key)",
        "description": "Return the value stored under ``key``, or ``null`` when the key does not exist.",
        "args": [("key", "string", False, "Key to look up.")],
        "returns": "string | null",
    },
    "globalThis.sessionStorage.setItem": {
        "sig": "sessionStorage.setItem(key, value)",
        "description": "Store ``value`` under ``key`` for the lifetime of the process.",
        "args": [("key", "string", False, "Key to write."), ("value", "string", False, "Value to store.")],
        "returns": "undefined",
    },
    "globalThis.sessionStorage.removeItem": {
        "sig": "sessionStorage.removeItem(key)",
        "description": "Delete ``key`` from the area.",
        "args": [("key", "string", False, "Key to delete.")],
        "returns": "undefined",
    },
    "globalThis.sessionStorage.clear": {
        "sig": "sessionStorage.clear()",
        "description": "Delete every key in the area.",
        "returns": "undefined",
    },
    "globalThis.sessionStorage.key": {
        "sig": "sessionStorage.key(index)",
        "description": "Return the key at ``index`` in insertion order, or ``null`` when out of range.",
        "args": [("index", "number", False, "Zero-based key index.")],
        "returns": "string | null",
    },
    "globalThis.sessionStorage.length": {
        "description": "The number of stored keys.",
    },
    "globalThis.Storage.prototype.getItem": {
        "sig": "storage.getItem(key)",
        "description": "Return the value stored under ``key``, or ``null`` when absent.",
        "args": [("key", "string", False, "Key to look up.")],
        "returns": "string | null",
    },
    "globalThis.Storage.prototype.setItem": {
        "sig": "storage.setItem(key, value)",
        "description": "Store a string value under ``key``.",
        "args": [("key", "string", False, "Key to write."), ("value", "string", False, "Value to store.")],
        "returns": "undefined",
    },
    "globalThis.Storage.prototype.removeItem": {
        "sig": "storage.removeItem(key)",
        "description": "Delete ``key`` from the area.",
        "args": [("key", "string", False, "Key to delete.")],
        "returns": "undefined",
    },
    "globalThis.Storage.prototype.clear": {
        "sig": "storage.clear()",
        "description": "Delete every key in the area.",
        "returns": "undefined",
    },
    "globalThis.Storage.prototype.key": {
        "sig": "storage.key(index)",
        "description": "Return the key at ``index`` in insertion order, or ``null`` when out of range.",
        "args": [("index", "number", False, "Zero-based key index.")],
        "returns": "string | null",
    },
    "globalThis.require": {
        "sig": "require(id)",
        "description": "Load a CommonJS builtin or file relative to the current module. In ``-e`` and stdin code the base directory is the working directory.",
        "args": [("id", "string", False, "Builtin name, package name, or relative/absolute path.")],
        "returns": "any",
    },
    "globalThis.os": {
        "description": "GodeJS exposes ``os`` as a convenience global in ``--eval`` code. It is the same object as ``require('node:os')``.",
    },
    "globalThis.global": {
        "description": "An alias for ``globalThis``, matching Node's CommonJS global.",
    },
}

MODULE_META = {
    "http3": {
        "title": "http3 (GodeJS extension)",
        "description": CURATED["http3"]["description"],
        "example": CURATED["http3"]["example"],
    },
}

HEADING_RE = re.compile(r"^(#{2,4})\s+(.+?)\s*$", re.M)
CODE_RE = re.compile(r"```(js|cjs|mjs|javascript)\n(.*?)```", re.S)
ANY_CODE_RE = re.compile(r"```.*?```", re.S)
PARAM_RE = re.compile(r"^\* `([^`]+)` \{([^}]+)\}(.*)$")


def strip_comments(text):
    return re.sub(r"<!--.*?-->", "", text, flags=re.S)


def clean_md(text):
    """Turn the small Markdown subset used by Node docs into RST."""
    text = strip_comments(text)
    text = re.sub(r"^>\s*(Stability|Added in|Deprecated|Status).*$", "", text, flags=re.M)
    text = re.sub(r"^\s*\[[^\]]+\]:\s+\S+.*$", "", text, flags=re.M)
    tokens = {}

    def stash(value):
        key = "\x00%d\x00" % len(tokens)
        tokens[key] = value
        return key

    text = re.sub(r"`{3,}(.*?)`{3,}", lambda m: stash("``%s``" % m.group(1).replace("`", "")), text, flags=re.S)
    text = re.sub(r"`([^`]+)`", lambda m: stash("``%s``" % m.group(1)), text)
    text = re.sub(r"(?<![\w])_([^_\n]+)_(?![\w])", r"*\1*", text)
    text = re.sub(r"\[`([^`]+)`\]\[[^\]]*\]", r"\1", text)
    text = re.sub(r"\[([^\]]+)\]\(([^)]+)\)",
                  lambda m: stash("`%s <%s>`_" % (m.group(1), m.group(2))), text)
    text = re.sub(r"\[([^\]]+)\]\[[^\]]*\]", r"\1", text)
    text = re.sub(r"\[([^\]]+)\]\[\]", r"\1", text)
    text = re.sub(r"\{([^{}]+)\}", lambda m: stash("``%s``" % m.group(1)), text)
    text = re.sub(r"\[\^[^\]]+\]", "", text)
    text = text.replace("\u00a0", " ")
    text = text.replace("_", "\\_")
    for key, value in tokens.items():
        text = text.replace(key, value)
    text = re.sub(r"`{3,}", "", text)
    return balance_inline_literals(text.strip())


def balance_inline_literals(text):
    """Escape backticks that RST would treat as unterminated markup."""
    out = []
    i = 0
    while i < len(text):
        if text.startswith("``", i):
            end = text.find("``", i + 2)
            if end < 0:
                out.append("\\`\\`")
                i += 2
                continue
            out.append(text[i:end + 2])
            i = end + 2
            if i < len(text) and text[i].isalnum():
                out.append("\\ ")
            continue
        if text[i] == "`":
            out.append("\\`")
            i += 1
            continue
        out.append(text[i])
        i += 1
    return "".join(out)


def flatten(text):
    text = clean_md(text)
    lines = []
    for line in text.splitlines():
        line = line.strip()
        if not line:
            continue
        lines.append(line)
    return " ".join(lines)


def paragraphs(text, limit=3):
    text = clean_md(text)
    chunks = re.split(r"\n\s*\n", text)
    out = []
    for chunk in chunks:
        chunk = " ".join(line.strip() for line in chunk.splitlines() if line.strip())
        if not chunk or chunk.startswith("```") or chunk.startswith("|"):
            continue
        out.append(chunk)
        if len(out) >= limit:
            break
    return "\n\n".join(out)


def parse_signature_args(sig):
    open_at = sig.find("(")
    close_at = sig.rfind(")")
    if open_at < 0 or close_at <= open_at:
        return []
    inner = sig[open_at + 1:close_at]
    parts, buf, depth = [], [], 0
    for char in inner:
        if char in "[{(":
            depth += 1
        elif char in "]})":
            depth = max(0, depth - 1)
        if char == "," and depth == 0:
            parts.append("".join(buf))
            buf = []
        else:
            buf.append(char)
    if "".join(buf).strip():
        parts.append("".join(buf))

    args = []
    for part in parts:
        part = part.strip()
        if not part:
            continue
        starts_optional = part.startswith("[")
        identifiers = re.findall(r"[A-Za-z_$][A-Za-z0-9_$]*", part)
        if not identifiers:
            continue
        main = None if starts_optional else identifiers[0]
        if main:
            args.append({
                "name": main,
                "optional": starts_optional,
                "variadic": part.startswith("...") or ("..." + main) in part,
            })
        for ident in identifiers[(0 if starts_optional else 1):]:
            if ident in ("new", "class", "function", "options", "true", "false"):
                pass
            if any(a["name"] == ident for a in args):
                continue
            args.append({
                "name": ident,
                "optional": True,
                "variadic": ("..." + ident) in part,
            })
    return args


def heading_signature(title):
    text = re.sub(r"^(Class|Static method|Static property|Method|Property)\s*:\s*", "", title)
    text = re.sub(r"^Event\s*:\s*", "", text)
    spans = re.findall(r"`([^`]+)`", text)
    if not spans:
        return None, None
    sig = spans[0].strip()
    sig = re.sub(r"^new\s+", "", sig)
    sig = sig.rstrip(".")
    name = sig.split("(", 1)[0].strip()
    name = re.sub(r"\s+", "", name)
    if not name or not re.match(r"^[A-Za-z_$@]", name):
        return None, None
    return sig, name


def normalize_name(name):
    name = name.replace("`", "").strip()
    name = name.replace("fsPromises.", "fs.promises.")
    name = name.replace("fsPromises", "fs.promises")
    name = name.replace("node:", "")
    name = re.sub(r"^new\s+", "", name)
    name = name.rstrip(".")
    return name


def section_symbol(name):
    return name.rsplit(".", 1)[-1]


def section_owner(name):
    return name.rsplit(".", 1)[0] if "." in name else ""


def extract_param_bullets(body):
    params = {}
    for line in body.splitlines():
        m = PARAM_RE.match(line)
        if not m:
            continue
        name, typ, desc = m.group(1), m.group(2), m.group(3).strip()
        params[name] = {"type": typ, "desc": flatten(desc)}
    return params


def extract_returns(body):
    m = re.search(r"^#### Returns\s*$(.*?)(?=^#### |\Z)", body, re.M | re.S)
    if not m:
        return ""
    return paragraphs(m.group(1), limit=1)


def extract_example(body):
    matches = CODE_RE.findall(body)
    if not matches:
        return ""
    priority = {"cjs": 0, "js": 1, "mjs": 2, "javascript": 3}
    matches.sort(key=lambda item: priority.get(item[0], 9))
    code = matches[0][1].strip("\n")
    if len(code) > 3200:
        code = code[:3200].rstrip() + "\n// ... truncated"
    return code


def description_from(body):
    body = strip_comments(body)
    cut = re.search(r"^#{3,4}\s+", body, flags=re.M)
    if cut:
        body = body[:cut.start()]
    body = ANY_CODE_RE.sub("", body)
    lines, skipping = [], False
    for line in body.splitlines():
        if PARAM_RE.match(line):
            skipping = True
            continue
        if skipping:
            if line.startswith(("  ", "\t")) or not line.strip():
                continue
            skipping = False
        if not any(part.strip() for part in lines) and line.lstrip().startswith("* "):
            skipping = True
            continue
        if line.startswith("> "):
            continue
        lines.append(line)
    return paragraphs("\n".join(lines), limit=2)


def parse_sections():
    sections = []
    if not os.path.isdir(NODE_DOCS):
        return sections
    for filename in sorted(os.listdir(NODE_DOCS)):
        if not filename.endswith(".md"):
            continue
        path = os.path.join(NODE_DOCS, filename)
        try:
            with open(path, encoding="utf-8") as f:
                text = f.read()
        except OSError:
            continue
        matches = list(HEADING_RE.finditer(text))
        for i, match in enumerate(matches):
            level = len(match.group(1))
            title = match.group(2)
            start = match.end()
            end = len(text)
            for later in matches[i + 1:]:
                if len(later.group(1)) <= level:
                    end = later.start()
                    break
            body = text[start:end]
            sig, name = heading_signature(title)
            if not sig:
                continue
            if "Event:" in title or title.startswith(("History", "Deprecated")):
                continue
            name = normalize_name(name)
            if not name:
                continue
            sections.append({
                "file": filename,
                "level": level,
                "title": title,
                "sig": sig,
                "name": name,
                "symbol": section_symbol(name),
                "owner": section_owner(name),
                "args": parse_signature_args(sig),
                "params": extract_param_bullets(body),
                "description": description_from(body),
                "returns": extract_returns(body),
                "example": extract_example(body),
                "anchor": anchor_for(title),
            })
    return sections


def anchor_for(title):
    text = title.replace("`", "").lower()
    text = re.sub(r"[^\w\s-]", "", text)
    return re.sub(r"\s+", "-", text).strip("-")


def parse_module_intro(filename):
    path = os.path.join(NODE_DOCS, filename)
    if not os.path.isfile(path):
        return "", "", ""
    with open(path, encoding="utf-8") as f:
        text = f.read()
    first_h2 = re.search(r"^##\s+", text, flags=re.M)
    intro = text[:first_h2.start()] if first_h2 else text
    intro = re.sub(r"^# .*$", "", intro, count=1, flags=re.M)
    intro = ANY_CODE_RE.sub("", intro)
    desc = paragraphs(intro, limit=2)
    blocks = parse_examples(text[:first_h2.start()] if first_h2 else text)
    return desc, blocks, ""


def parse_examples(text):
    matches = CODE_RE.findall(text)
    if not matches:
        return ""
    priority = {"cjs": 0, "js": 1, "mjs": 2, "javascript": 3}
    matches.sort(key=lambda item: priority.get(item[0], 9))
    return matches[0][1].strip("\n")


def runtime_rows(api, module):
    rows = []
    seen = set()
    for row in api.get(module, []):
        if row.get("k") in ("ERROR", "getter-error"):
            continue
        path = row["p"]
        if path in seen:
            continue
        seen.add(path)
        rows.append(row)
    rows.sort(key=lambda row: row["p"].lower())
    return rows


def runtime_short(module, path):
    if module == "__globals__":
        short = path
        if short.startswith("globalThis."):
            short = short[len("globalThis."):]
    elif path.startswith(module + "."):
        short = path[len(module) + 1:]
    else:
        short = path
    short = re.sub(r"\.prototype\.", ".", short)
    short = re.sub(r"\(\)$", "", short)
    return short


def match_section(module, short, sections, index, arity):
    if module != "__globals__":
        doc_file = DOC_FILES.get(module)
        if not doc_file or not os.path.isfile(os.path.join(NODE_DOCS, doc_file)):
            return None
    exact_keys = [module + "." + short, short]
    normalized_short = normalize_name(short)
    for key in exact_keys:
        for sec in index["exact"].get(normalize_name(key), []):
            if module == "__globals__" or sec["file"] == DOC_FILES.get(module) or sec["name"].endswith("." + short):
                return sec
    best, best_score = None, 0
    symbol = short.rsplit(".", 1)[-1]
    owner = short.rsplit(".", 1)[0].rsplit(".", 1)[-1] if "." in short else ""
    for sec in index["symbol"].get(symbol, []):
        score = 40
        if sec["file"] == DOC_FILES.get(module):
            score += 20
        if owner and sec["owner"].lower() == owner.lower():
            score += 25
        if len(sec["args"]) == arity:
            score += 10
        if sec["name"].lower().endswith("." + normalized_short.lower()):
            score += 30
        if score > best_score:
            best, best_score = sec, score
    return best if best_score >= 55 else None


def curated_for(module, short):
    keys = [module + "." + short, short]
    if module == "__globals__":
        keys.append("globalThis." + short)
        if short.startswith("Storage."):
            keys.append("globalThis.Storage.prototype." + short.split(".", 1)[1])
    for key in keys:
        if key in CURATED:
            return CURATED[key]
    return None


def module_marker(module):
    if module == "http3":
        return "GodeJS extension"
    if module == "test":
        return "Node test harness"
    return ""


def group_for(module, path):
    if module == "__globals__":
        short = path[len("globalThis."):] if path.startswith("globalThis.") else path
    elif path.startswith(module + "."):
        short = path[len(module) + 1:]
    else:
        short = path
    if ".prototype." in short:
        return short.split(".prototype.", 1)[0]
    parts = short.split(".")
    if len(parts) > 1:
        return parts[0]
    return "Top level"


def group_title(group):
    if group == "Top level":
        return "Top level"
    return group + " class" if group and group[0].isupper() else group


def rst_heading(text, char):
    text = re.sub(r"\s+", " ", text).strip()
    return text + "\n" + char * max(len(text), 3) + "\n"


def signature_for(module, short, row):
    if module == "__globals__":
        name = "globalThis." + short
    else:
        name = module + "." + short
    arity = int(row.get("a") or 0)
    args = []
    if arity > 0:
        args = ["arg%d" % i for i in range(arity)]
    else:
        args = ["...args"]
    return "%s(%s)" % (name, ", ".join(args))


def format_arg(arg, sec):
    info = None
    if sec:
        info = sec["params"].get(arg["name"])
    arg_type = info["type"] if info else arg.get("type")
    arg_desc = info["desc"] if info else arg.get("desc")
    label = "``%s``" % arg["name"]
    if arg.get("variadic"):
        label = "``...%s``" % arg["name"]
    if arg_type:
        label += " (*%s*)" % arg_type
    if arg.get("optional"):
        label += " — optional"
    if arg_desc:
        label += ": " + arg_desc
    return "* " + label


def render_entry(module, row, sec, curated):
    short = runtime_short(module, row["p"])
    kind = row["k"]
    marker = module_marker(module)
    title = short.replace("`", "")
    out = [rst_heading(title, "~")]

    sig = None
    args = []
    description = ""
    returns = ""
    example = ""
    if curated:
        sig = curated.get("sig")
        args = [{"name": n, "type": t, "optional": o, "desc": d}
                for n, t, o, d in curated.get("args", [])]
        description = curated.get("description", "")
        returns = curated.get("returns", "")
        example = curated.get("example", "")
    elif sec:
        sig = sec["sig"]
        args = sec["args"]
        description = sec["description"]
        returns = sec["returns"]
        example = sec["example"]
    if not sig:
        sig = signature_for(module, short, row)

    is_class = bool(sec and sec["title"].startswith("Class:"))
    out.append(".. code-block:: text\n")
    out.append("   " + sig + "\n")
    shown_kind = "class" if is_class else KIND_LABEL.get(kind, kind)
    arity = len(args) if args else int(row.get("a") or 0)
    line = "*Kind:* %s · *Arguments:* %d" % (shown_kind, arity)
    if marker:
        line += " · *%s*" % marker
    out.append(line + "\n")
    if description:
        out.append(description + "\n")
    if args:
        out.append("**Arguments**\n")
        for arg in args:
            if isinstance(arg, dict):
                out.append(format_arg(arg, sec))
            else:
                out.append("* ``%s``" % arg)
        out.append("")
    elif is_class:
        out.append("This is a class; construct instances with ``new``.\n")
    elif sec:
        out.append("This callable takes no arguments.\n")
    if returns:
        out.append("**Returns**\n")
        out.append(returns + "\n")
    if example:
        out.append("**Example**\n")
        out.append(".. code-block:: javascript\n")
        for line in example.splitlines():
            out.append("   " + line if line else "")
        out.append("")
    if sec:
        url = "https://nodejs.org/docs/latest-v24.x/api/%s#%s" % (
            sec["file"].replace(".md", ".html"), sec["anchor"])
        out.append("`Node.js reference <%s>`__" % url)
    elif not curated:
        out.append(".. note::\n")
        out.append("   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.\n")
    elif module == "http3":
        out.append("`GodeJS HTTP/3 guide <../../http3.html>`__")
    out.append("")
    return "\n".join(out)


def render_group_intro(module, group, rows, sections, index):
    if group == "Top level":
        return ""
    short = group
    arity = 0
    sec = match_section(module, short, sections, index, arity)
    curated = curated_for(module, short)
    if curated and curated.get("description"):
        return "**%s** — %s\n" % (group, curated["description"])
    if sec and sec["description"]:
        return "**%s** — %s\n" % (group, sec["description"])
    return ""


def render_module(module, rows, sections, index):
    title = TITLES.get(module, module)
    meta = MODULE_META.get(module, {})
    title = meta.get("title", title)
    out = [rst_heading(title, "=")]
    description = meta.get("description", "")
    example = meta.get("example", "")
    if module in DOC_FILES:
        intro, intro_example, _ = parse_module_intro(DOC_FILES[module])
        description = description or intro
        example = example or intro_example
    if not description:
        description = "The ``%s`` builtin." % module
    out.append(description + "\n")

    if module == "http3":
        loading = "const http3 = require('gode:http3');"
        esm = "import http3 from 'gode:http3';"
    else:
        loading = "const %s = require('node:%s');" % (module.replace("/", "_"), module)
        esm = "import * as %s from 'node:%s';" % (module.replace("/", "_"), module)
    out.append("Loading\n-------\n")
    out.append(".. code-block:: javascript\n")
    out.append("   " + loading)
    out.append("   // ESM")
    out.append("   " + esm + "\n")
    if example:
        out.append("Quick start\n-----------\n")
        out.append(".. code-block:: javascript\n")
        for line in example.splitlines():
            out.append("   " + line if line else "")
        out.append("")

    functions = [r for r in rows if r["k"] == "function"]
    values = [r for r in rows if r["k"] != "function"]
    if functions:
        out.append("API\n---\n")
        groups = {}
        for row in functions:
            groups.setdefault(group_for(module, row["p"]), []).append(row)
        order = sorted(groups, key=lambda g: (g != "Top level", g.lower()))
        for group in order:
            if group != "Top level":
                out.append(rst_heading(group_title(group), "-"))
                intro = render_group_intro(module, group, groups[group], sections, index)
                if intro:
                    out.append(intro)
            for row in groups[group]:
                short = runtime_short(module, row["p"])
                sec = match_section(module, short, sections, index, int(row.get("a") or 0))
                curated = curated_for(module, short)
                out.append(render_entry(module, row, sec, curated))
    if values:
        out.append("Properties and constants\n------------------------\n")
        for row in values:
            short = runtime_short(module, row["p"])
            out.append("* ``%s`` — %s" % (short, KIND_LABEL.get(row["k"], row["k"])))
        out.append("")
    out.append(".. note::\n")
    out.append("   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.\n")
    return "\n".join(out)


def render_globals(api, sections, index):
    rows = runtime_rows(api, "__globals__")
    out = [rst_heading("Global objects", "=")]
    out.append("GodeJS installs the Node.js host globals plus the Web APIs it implements. Standard ECMAScript built-ins (``Object``, ``Array``, ``Promise``, ``Map``, typed arrays, and so on) are not repeated here; use MDN or the ECMAScript specification for those.\n")
    groups = {}
    top_functions = []
    for row in rows:
        short = runtime_short("__globals__", row["p"])
        if "." not in short and row["k"] == "function":
            sec = match_section("__globals__", short, sections, index, int(row.get("a") or 0))
            if not (sec and sec["title"].startswith("Class:")):
                top_functions.append(row)
                continue
        root = short.split(".", 1)[0]
        groups.setdefault(root, []).append(row)
    if top_functions:
        out.append(rst_heading("Functions", "-"))
        for row in top_functions:
            short = runtime_short("__globals__", row["p"])
            sec = match_section("__globals__", short, sections, index, int(row.get("a") or 0))
            curated = curated_for("__globals__", short)
            out.append(render_entry("__globals__", row, sec, curated))
    order = sorted(groups, key=lambda g: (g.lower() != "localstorage", g.lower()))
    for root in order:
        group_rows = groups[root]
        out.append(rst_heading(root, "-"))
        root_row = next((r for r in group_rows if runtime_short("__globals__", r["p"]) == root), None)
        if root_row:
            sec = match_section("__globals__", root, sections, index, 0)
            curated = curated_for("__globals__", root)
            if curated and curated.get("description"):
                out.append(curated["description"] + "\n")
            elif sec and sec["description"]:
                out.append(sec["description"] + "\n")
        funcs = [r for r in group_rows if r["k"] == "function" and r is not root_row]
        values = [r for r in group_rows if r["k"] != "function" and runtime_short("__globals__", r["p"]) != root]
        for row in funcs:
            short = runtime_short("__globals__", row["p"])
            sec = match_section("__globals__", short, sections, index, int(row.get("a") or 0))
            curated = curated_for("__globals__", short)
            out.append(render_entry("__globals__", row, sec, curated))
        if values:
            out.append("**Properties**\n")
            for row in values:
                short = runtime_short("__globals__", row["p"])
                out.append("* ``%s`` — %s" % (short, KIND_LABEL.get(row["k"], row["k"])))
            out.append("")
    return "\n".join(out)


def main():
    with open(API_JSON, encoding="utf-8") as f:
        api = json.load(f)
    os.makedirs(OUT, exist_ok=True)
    sections = parse_sections()
    index = {"exact": {}, "symbol": {}}
    for sec in sections:
        index["exact"].setdefault(sec["name"], []).append(sec)
        index["symbol"].setdefault(sec["symbol"], []).append(sec)

    pages, callables, matched, examples = [], 0, 0, 0
    for module in TITLES:
        rows = runtime_rows(api, module)
        if not rows:
            continue
        text = render_module(module, rows, sections, index)
        with open(os.path.join(OUT, module.replace("/", "-") + ".rst"), "w", encoding="utf-8") as f:
            f.write(text)
        pages.append(module)
        for row in rows:
            if row["k"] != "function":
                continue
            callables += 1
            short = runtime_short(module, row["p"])
            sec = match_section(module, short, sections, index, int(row.get("a") or 0))
            if sec:
                matched += 1
                if sec["example"]:
                    examples += 1

    globals_rows = runtime_rows(api, "__globals__")
    with open(os.path.join(OUT, "globals.rst"), "w", encoding="utf-8") as f:
        f.write(render_globals(api, sections, index))
    for row in globals_rows:
        if row["k"] != "function":
            continue
        callables += 1
        short = runtime_short("__globals__", row["p"])
        sec = match_section("__globals__", short, sections, index, int(row.get("a") or 0))
        if sec:
            matched += 1
            if sec["example"]:
                examples += 1

    index_lines = [rst_heading("API Reference", "=")]
    index_lines.append("This reference is generated from the running runtime, not hand-maintained. Every callable the runtime exposes has its own entry with a signature, typed arguments, return value, and an example where the upstream Node.js documentation provides one. Private helpers and GodeJS extensions are labelled explicitly.\n")
    index_lines.append("Signatures and prose are derived from the Node.js v24.22.2 documentation, licensed under the MIT License. The runtime surface is the source of truth for what is available here.\n")
    index_lines.append("Global objects\n--------------\n")
    index_lines.append(".. toctree::\n   :maxdepth: 1\n\n   globals\n")
    for title, modules in GROUPS:
        index_lines.append(title + "\n" + "-" * len(title) + "\n")
        index_lines.append(".. toctree::\n   :maxdepth: 1\n")
        for module in modules:
            if module in pages:
                index_lines.append("   " + module)
        index_lines.append("")
    with open(os.path.join(OUT, "index.rst"), "w", encoding="utf-8") as f:
        f.write("\n".join(index_lines))

    stale = {module.replace("/", "-") + ".rst" for module in pages}
    stale.update({"globals.rst", "index.rst"})
    for name in os.listdir(OUT):
        if name.endswith(".rst") and name not in stale:
            os.remove(os.path.join(OUT, name))

    print("modules: %d, callables: %d, matched: %d, examples: %d, doc sections: %d" %
          (len(pages), callables, matched, examples, len(sections)), file=sys.stderr)


if __name__ == "__main__":
    main()
