#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Read version from ota.go
const otaPath = path.join(__dirname, '../../ota.go');
const otaContent = fs.readFileSync(otaPath, 'utf8');
const versionMatch = otaContent.match(/var builtAppVersion = "(.+)"/);

if (!versionMatch) {
  console.error('Failed to extract version from ota.go');
  process.exit(1);
}

let version = versionMatch[1];

// Remove +dev suffix for Tauri (Tauri doesn't support + in versions)
version = version.replace('+dev', '-dev');

console.log(`Extracted version: ${version}`);

// Update tauri.conf.json
const tauriConfPath = path.join(__dirname, '../src-tauri/tauri.conf.json');
const tauriConf = JSON.parse(fs.readFileSync(tauriConfPath, 'utf8'));
tauriConf.version = version;
fs.writeFileSync(tauriConfPath, JSON.stringify(tauriConf, null, 2) + '\n');

console.log(`✓ Updated tauri.conf.json to version ${version}`);

// Update package.json
const packagePath = path.join(__dirname, '../package.json');
const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
packageJson.version = version;
fs.writeFileSync(packagePath, JSON.stringify(packageJson, null, 2) + '\n');

console.log(`✓ Updated package.json to version ${version}`);
