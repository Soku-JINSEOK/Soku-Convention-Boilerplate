import gts from 'gts';

export default [
  {
    ignores: ['**/dist/**', '**/coverage/**', '**/node_modules/**'],
  },
  ...gts,
];
