import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

const forms = {
  'bug_report.yml': ['defect', 'reproduce', 'evidence', 'acceptance'],
  'chore_task.yml': ['goal', 'work', 'acceptance'],
  'docs_update.yml': ['goal', 'content', 'acceptance'],
  'feature_request.yml': ['goal', 'scope', 'acceptance'],
  'refactor_proposal.yml': ['rationale', 'design', 'acceptance'],
};

for (const [file, taskFields] of Object.entries(forms)) {
  test(`${file} keeps task fields after English and AI Assistance last`, () => {
    const path = fileURLToPath(new URL(`./ISSUE_TEMPLATE/${file}`, import.meta.url));
    const source = readFileSync(path, 'utf8');
    const ids = [...source.matchAll(/^\s+id:\s+(\S+)\s*$/gm)].map((match) => match[1]);
    assert.deepEqual(ids, [
      ...taskFields,
      'priority',
      'area',
      'safety',
      'summary_ko',
      'summary_ja',
      'ai_assistance',
    ]);

    for (const id of ids) {
      const start = source.indexOf(`  id: ${id}`);
      const next = source.indexOf('\n  - type:', start);
      const field = source.slice(start, next < 0 ? source.length : next);
      assert.match(field, /validations:\n\s+required: true/, `${id} must be required`);
    }
  });
}
