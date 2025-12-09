#!/usr/bin/env node
const fs = require("fs");
const path = require("path");
const patterns = require("./patterns");

const ROOT = process.cwd();
const IGNORE = new Set(["node_modules", ".git"]);

function walk(dir, fn) {
  for (const item of fs.readdirSync(dir)) {
    const full = path.join(dir, item);
    const stat = fs.lstatSync(full);
    if (stat.isDirectory()) {
      if (!IGNORE.has(item)) walk(full, fn);
    } else {
      fn(full);
    }
  }
}

const findings = [];
walk(ROOT, file => {
  const content = fs.readFileSync(file, "utf8");

  for (const pattern of patterns.secrets) {
    let match;
    while ((match = pattern.re.exec(content)) !== null) {
      findings.push({
        type: pattern.name,
        file,
        match: match[0].slice(0, 100)
      });
    }
  }
});

fs.writeFileSync("scan-report.json", JSON.stringify({ findings }, null, 2));
console.log("Scan complete. Findings:", findings.length);
