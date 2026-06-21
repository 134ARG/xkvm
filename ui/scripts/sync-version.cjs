#!/usr/bin/env node

// Writes the given version into package.json and tauri.conf.json.
// The version is passed as argv[2] by bump_version.sh (the single source of
// truth). When run standalone, it falls back to packaging/version.txt.

const fs = require('fs');
const path = require('path');

function npmVersion() {
  if (process.argv[2]) return process.argv[2];
  // Fallback: derive from canonical packaging/version.txt (+dev -> -dev).
  const txt = path.join(__dirname, '../../packaging/version.txt');
  return fs.readFileSync(txt, 'utf8').trim().replace('+dev', '-dev');
}

const version = npmVersion();
console.log(`Syncing JSON files to version: ${version}`);

for (const rel of ['../src-tauri/tauri.conf.json', '../package.json']) {
  const file = path.join(__dirname, rel);
  const json = JSON.parse(fs.readFileSync(file, 'utf8'));
  json.version = version;
  fs.writeFileSync(file, JSON.stringify(json, null, 2) + '\n');
  console.log(`✓ Updated ${path.basename(file)} to ${version}`);
}
