#!/usr/bin/env node

import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const version = readFileSync(new URL('../VERSION', import.meta.url), 'utf8').trim();
const zimaos = readFileSync(new URL('../zimaos-compose.yaml', import.meta.url), 'utf8');
const composeFiles = [
  readFileSync(new URL('../docker-compose.yml', import.meta.url), 'utf8'),
  readFileSync(new URL('../docker-compose.example.yml', import.meta.url), 'utf8'),
];

assert.match(version, /^\d+\.\d+\.\d+$/, 'VERSION must be semantic');
assert.doesNotMatch(zimaos, /\$\{/, 'primary ZimaOS Compose must not interpolate variables');

const imageMatch = zimaos.match(
  /^\s*image:\s*ghcr\.io\/maroishiku\/dyniku:([^@\s]+)@(sha256:[a-f0-9]{64})\s*$/m,
);
assert.ok(imageMatch, 'ZimaOS image must use a direct version and digest');
assert.equal(imageMatch[1], version, 'ZimaOS image version must match VERSION');

const expectedImage = `ghcr.io/maroishiku/dyniku:${version}@${imageMatch[2]}`;
for (const compose of composeFiles) {
  assert.ok(compose.includes(`\${DYNIKU_IMAGE:-${expectedImage}}`), 'Compose fallback must use the release version and digest');
  assert.match(compose, /["']65000:8507["']/, 'Compose must publish host port 65000 to container port 8507');
}

const secretMatch = zimaos.match(/^\s*- ISHIKU_SETUP_SECRET=(.+)$/m);
assert.ok(secretMatch, 'ZimaOS must expose the setup secret directly');
assert.match(secretMatch[1], /^REPLACE-WITH-/, 'ZimaOS must ship only a synthetic setup-secret placeholder');
assert.ok(secretMatch[1].length >= 32, 'ZimaOS setup-secret placeholder must have at least 32 characters');
assert.doesNotMatch(zimaos, /ISHIKU_SETUP_SECRET_FILE|\/run\/secrets/, 'ZimaOS must not depend on an external setup-secret file');

assert.match(zimaos, /target:\s*8507/, 'ZimaOS container port must remain 8507');
assert.match(zimaos, /published:\s*"65000"/, 'ZimaOS host port must be 65000');
assert.match(zimaos, /port_map:\s*"65000"/, 'ZimaOS metadata must advertise host port 65000');

process.stdout.write(`ZimaOS Compose policy verified for Dyniku ${version} (${imageMatch[2]}).\n`);
