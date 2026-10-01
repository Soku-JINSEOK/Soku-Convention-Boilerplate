#!/usr/bin/env node
// Text guard only: grammar, rendering and implementation parity need review.
import {execFileSync} from 'node:child_process';
import {readFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

export function diagramTexts(path, source) {
  if (path.endsWith('.mmd')) return [source];
  if (path.endsWith('.svg')) {
    return [...source.matchAll(/<(text|title|desc)\b[^>]*>([\s\S]*?)<\/\1>/g)]
      .map((match) => match[2].replace(/<[^>]*>/g, ''))
      .concat([...source.matchAll(/\baria-label=["']([^"']*)["']/g)]
        .map((match) => match[1]));
  }
  if (!path.endsWith('.md')) return [];
  const openingPattern = new RegExp('^ {0,3}(' + String.fromCharCode(96) + '{3,}|~{3,})(.*)$');
  const texts = [];
  let fence = null;
  let diagram = false;
  let lines = [];
  for (const line of source.split(/\r?\n/)) {
    if (fence === null) {
      const opening = line.match(openingPattern);
      if (!opening) continue;
      fence = opening[1];
      diagram = opening[2].trim() === 'mermaid';
      lines = [];
    } else if (new RegExp('^ {0,3}' + fence[0] + '{' + fence.length + ',}\\s*$').test(line)) {
      if (diagram) texts.push(lines.join('\n'));
      fence = null;
    } else if (diagram) {
      lines.push(line);
    }
  }
  if (fence !== null && diagram) throw new Error(path + ': unclosed Mermaid fence');
  return texts;
}

export function hasNonEnglishLetters(text) {
  const decoded = text.replace(/&#(x[0-9a-f]+|\d+);/gi, (_, value) => {
    const code = value[0].toLowerCase() === 'x'
      ? Number.parseInt(value.slice(1), 16) : Number(value);
    return code <= 0x10ffff ? String.fromCodePoint(code) : '\ufffd';
  });
  return [...decoded].some((char) => char.codePointAt(0) > 127 && /\p{Letter}/u.test(char));
}

export function verifyDiagramText() {
  const paths = execFileSync('git', ['ls-files', '-z'], {encoding: 'utf8'})
    .split('\0').filter((path) => /\.(md|mmd|svg)$/.test(path));
  let count = 0;
  const failures = [];
  for (const path of paths) {
    const texts = diagramTexts(path, readFileSync(path, 'utf8'));
    count += texts.length;
    if (texts.some(hasNonEnglishLetters)) failures.push(path);
  }
  if (failures.length) throw new Error('Use English diagram text: ' + failures.join(', '));
  console.log('Diagram text guard passed for ' + count + ' text blocks; semantic review remains required.');
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  verifyDiagramText();
}
