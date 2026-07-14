export const TITLE_CONVENTIONS = Object.freeze([
  ['✨', 'feat'],
  ['🐛', 'fix'],
  ['♻️', 'refactor'],
  ['🎨', 'style'],
  ['📚', 'docs'],
  ['✅', 'test'],
  ['🔧', 'chore'],
  ['🚀', 'perf'],
  ['📦', 'build'],
  ['👷', 'ci'],
  ['🔥', 'remove'],
  ['🚑', 'hotfix'],
  ['🔖', 'release'],
  ['🔄', 'sync'],
  ['🔒️', 'security'],
  ['⏪️', 'revert'],
]);

const SCOPE_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const ENGLISH_SUBJECT_PATTERN = /^[\x20-\x7e]+$/;

export function validateContributionTitle(title) {
  const value = title.trim();
  const convention = TITLE_CONVENTIONS.find(([emoji, type]) =>
    value.startsWith(`${emoji} ${type}(`),
  );

  if (!convention) {
    return {
      valid: false,
      message:
        'Title must start with a supported Gitmoji/type pair, such as ' +
        '`📚 docs(workflow): ...`.',
    };
  }

  const [emoji, type] = convention;
  const remainder = value.slice(`${emoji} ${type}(`.length);
  const separator = remainder.indexOf('): ');
  if (separator < 1) {
    return {
      valid: false,
      message: 'Title must use `<gitmoji> <type>(<scope>): <English subject>`.',
    };
  }

  const scope = remainder.slice(0, separator);
  const subject = remainder.slice(separator + 3);
  if (!SCOPE_PATTERN.test(scope)) {
    return {
      valid: false,
      message: 'Scope is required and must use lowercase kebab-case.',
    };
  }
  if (!subject || subject !== subject.trim()) {
    return {valid: false, message: 'English subject is required.'};
  }
  if (!ENGLISH_SUBJECT_PATTERN.test(subject)) {
    return {valid: false, message: 'Subject must be written in English.'};
  }

  return {valid: true, message: 'Title follows the repository convention.'};
}
