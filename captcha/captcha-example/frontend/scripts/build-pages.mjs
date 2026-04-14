import { resolve } from 'node:path'
import { build } from 'vite'
import { viteSingleFile } from 'vite-plugin-singlefile'

const root = process.cwd()
const outDir = resolve(root, '../internal/captcha/html')
const pages = [
  'odd-grid',
  'reality-swipe',
  'foreign-letter',
  'two-baskets',
  'track-object',
]

for (const [index, page] of pages.entries()) {
  await build({
    root,
    configFile: false,
    publicDir: false,
    plugins: [
      viteSingleFile({
        useRecommendedBuildConfig: false,
      }),
    ],
    build: {
      outDir,
      emptyOutDir: index === 0,
      chunkSizeWarningLimit: 1000,
      cssCodeSplit: false,
      modulePreload: false,
      target: 'es2020',
      rollupOptions: {
        input: resolve(root, `${page}.html`),
        output: {
          inlineDynamicImports: true,
        },
      },
    },
  })
}
