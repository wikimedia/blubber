import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Blubber",
  description: "Blubber Documentation",
  base: '/releng/blubber/',
  rewrites: {
    'README.md': 'index.md',
  },
  srcExclude: [
    'api',
    'cmd',
    'docker',
    'config',
    'examples/*/README.md',
    'meta',
    'out',
    'util',
    'scripts',
    'build',
    'buildkit',
    '**/TODO.md',
    '**/DEPENDENCIES.md'
  ],
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
    ],
    logo: '/logo-400.png',
    docFooter: {
      prev: false,
      next: false
    },
    sidebar: [
      {
        text: 'Documentation',
        items: [
          { text: 'Home',
            link: '/',
            items: [
              {text: 'Examples', link: '/#examples'},
              {text: 'Concepts', link: '/#concepts'},
              {text: 'Usage', link: '/#usage'}
            ]
          },
          { text: 'Configuration',
            link: '/configuration',
            items: [
              {text: 'Variants', link: '/configuration#variants'},
              {text: 'APT', link: '/configuration#apt'},
              {text: 'NodeJS', link: '/configuration#node-1'},
              {text: 'PHP', link: '/configuration#php-1'},
              {text: 'Python', link: '/configuration#python-1'}
            ]
          },
          { text: 'Examples',
            items: [
              {text: 'Basic usage', link: '/examples/01-basic-usage'},
              {text: 'Define environment', link: '/examples/02-defining-the-environment'},
              {text: 'Install packages', link: '/examples/03-installing-packages'},
              {text: 'Define builders', link: '/examples/04-builders'},
              {text: 'Copying from variants', link: '/examples/05-copying-from-other-variants'},
              {text: 'Python builder', link: '/examples/06-python-builder'},
              {text: 'NodeJS builder', link: '/examples/07-node-builder'},
              {text: 'Build arguments', link: '/examples/08-build-arguments'}
            ]
          }
        ],
      },
      {
        text: 'Development',
        items: [
          { text: 'Changelog', link: '/CHANGELOG'},
          { text: 'Code', link: 'https://gitlab.wikimedia.org/repos/releng/blubber'},
          { text: 'Contributing', link: '/CONTRIBUTING'},
          { text: 'Release', link: '/RELEASE'}
        ]
      }
    ],
    footer: {
      copyright: 'For copyright, licensing information, and website source code, see the <a href="https://gitlab.wikimedia.org/repos/releng/blubber">GitLab repository</a>.'
    },
    search: {
      provider: 'local'
    }
  }
})
