import assert from 'node:assert/strict';
import test from 'node:test';
import {diagramTexts, hasNonEnglishLetters} from './verify-diagram-text.mjs';

const fence = String.fromCharCode(96).repeat(3);

test('localized prose is allowed while Mermaid labels are inspected', () => {
  const source = '# 한국어\n\n' + fence + 'mermaid\nflowchart TD\n a["English"] --> b["한국어"]\n' + fence;
  const texts = diagramTexts('README.ko.md', source);
  assert.equal(texts.length, 1);
  assert.equal(hasNonEnglishLetters(texts[0]), true);
  assert.equal(hasNonEnglishLetters('English labels — punctuation is allowed'), false);
});

test('tilde fences work and Markdown examples are not mistaken for diagrams', () => {
  assert.deepEqual(diagramTexts('guide.md', '~~~mermaid\na["English"]\n~~~'), ['a["English"]']);
  const example = fence + String.fromCharCode(96) + 'markdown\n' + fence + 'mermaid\n日本語\n' + fence + '\n' + fence + String.fromCharCode(96);
  assert.deepEqual(diagramTexts('guide.md', example), []);
  assert.throws(() => diagramTexts('guide.md', fence + 'mermaid\nincomplete'), /unclosed/);
});

test('standalone sources and nested SVG text include encoded labels', () => {
  assert.equal(hasNonEnglishLetters(diagramTexts('map.mmd', 'a["日本語"]')[0]), true);
  const svg = '<svg aria-label="English"><title>English</title><text><tspan>&#xD55C;</tspan></text></svg>';
  assert.equal(diagramTexts('map.svg', svg).some(hasNonEnglishLetters), true);
  assert.deepEqual(diagramTexts('code.js', '한국어'), []);
});
