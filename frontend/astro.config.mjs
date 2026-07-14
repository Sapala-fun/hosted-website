import { defineConfig } from 'astro/config';
import githubPages from '@astrojs/github-pages';

export default defineConfig({
  site: 'https://www.sapala.fun',
  adapter: githubPages(),
});
