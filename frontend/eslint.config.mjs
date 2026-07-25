import js from '@eslint/js';
import configPrettier from 'eslint-config-prettier';
import typescriptParser from '@typescript-eslint/parser';
import typescriptPlugin from '@typescript-eslint/eslint-plugin';

export default [
  js.configs.recommended,
  configPrettier,
  {
    files: ['src/**/*.{js,ts}'],
    languageOptions: {
      parser: typescriptParser,
      globals: { console: 'readonly', exports: 'readonly', module: 'readonly' },
    },
    plugins: {
      '@typescript-eslint': typescriptPlugin,
    },
    rules: {
      ...typescriptPlugin.configs.recommended.rules,
      'no-console': 'off',
      'no-undef': 'off',
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^[_]' },
      ],
      '@typescript-eslint/triple-slash-reference': 'off',
    },
  },
];
