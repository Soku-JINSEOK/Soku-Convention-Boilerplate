export const TITLE_CONVENTIONS = Object.freeze([
  ['✨', 'feat'],
  ['🐛', 'fix'],
  ['📚', 'docs'],
  ['✅', 'test'],
  ['🔧', 'chore'],
  ['🚀', 'perf'],
  ['📦', 'build'],
  ['👷', 'ci'],
  ['🔄', 'sync'],
  ['🔒️', 'security'],
]);

const SCOPE_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const ENGLISH_SUBJECT_PATTERN = /^[\x20-\x7e]+$/;

export function validateContributionTitle(title) {
  const value = title.trim();
  let isBreaking = false;
  let convention = null;
  let prefixLength = 0;

  // 1. Check for breaking change pattern (starts with 💥 and ends type with !)
  if (value.startsWith('💥 ')) {
    const afterEmoji = value.slice('💥 '.length); // Remove '💥 '
    const matchedType = TITLE_CONVENTIONS.find(([_, type]) =>
      afterEmoji.startsWith(`${type}!(`),
    );
    if (matchedType) {
      isBreaking = true;
      const type = matchedType[1];
      convention = ['💥', type];
      prefixLength = `💥 ${type}!(`.length;
    }
  }

  // 2. If not breaking, check for standard prefix
  if (!convention) {
    convention = TITLE_CONVENTIONS.find(([emoji, type]) =>
      value.startsWith(`${emoji} ${type}(`),
    );
    if (convention) {
      const [emoji, type] = convention;
      prefixLength = `${emoji} ${type}(`.length;
    }
  }

  if (!convention) {
    return {
      valid: false,
      message:
        'Title must start with a supported Gitmoji/type pair, such as ' +
        '`📚 docs(workflow): ...`. For breaking changes, use `💥 feat!(scope): ...`.',
    };
  }

  const remainder = value.slice(prefixLength);
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
  if (subject.endsWith('.')) {
    return {valid: false, message: 'Subject must not end with a period.'};
  }
  if (value.length > 72) {
    return {
      valid: false,
      message: `Title exceeds 72 characters (currently ${value.length}).`,
    };
  }

  return {valid: true, message: 'Title follows the repository convention.'};
}
