import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

const forms = {
  'bug_report.yml': ['defect', 'reproduce', 'evidence'],
  'chore_task.yml': ['goal', 'work'],
  'docs_update.yml': ['goal', 'content'],
  'feature_request.yml': ['goal', 'scope'],
  'refactor_proposal.yml': ['rationale', 'design'],
};

for (const [file, taskFields] of Object.entries(forms)) {
  test(`${file} keeps required task fields and optional metadata policy`, () => {
    const path = fileURLToPath(new URL(`./ISSUE_TEMPLATE/${file}`, import.meta.url));
    const source = readFileSync(path, 'utf8');
    const ids = [...source.matchAll(/^\s+id:\s+(\S+)\s*$/gm)].map((match) => match[1]);
    const required = [...taskFields, 'acceptance', 'safety'];
    const optional = ['priority', 'area', 'summary_ko', 'summary_ja', 'ai_assistance'];
    const expected = [...required, ...optional];
    assert.deepEqual(ids.toSorted(), expected.toSorted());

    for (const id of required) {
      const start = source.indexOf(`  id: ${id}`);
      const next = source.indexOf('\n  - type:', start);
      const field = source.slice(start, next < 0 ? source.length : next);
      assert.match(
        field,
        /validations:\n\s+required: true/,
        `${id} must be required`,
      );
    }
    for (const id of optional) {
      const start = source.indexOf(`  id: ${id}`);
      const next = source.indexOf('\n  - type:', start);
      const field = source.slice(start, next < 0 ? source.length : next);
      assert.match(
        field,
        /validations:\n\s+required: false/,
        `${id} must be optional`,
      );
    }
  });
}
