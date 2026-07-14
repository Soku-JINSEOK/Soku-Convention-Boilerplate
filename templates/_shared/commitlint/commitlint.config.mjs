import {validateContributionTitle} from './contribution-title.mjs';

export default {
  plugins: [
    {
      rules: {
        'boilerplate-title': ({header}) => {
          const result = validateContributionTitle(header ?? '');
          return [result.valid, result.message];
        },
      },
    },
  ],
  rules: {
    'boilerplate-title': [2, 'always'],
  },
};
