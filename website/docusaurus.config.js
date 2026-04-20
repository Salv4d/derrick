// @ts-check
import {themes as prismThemes} from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Derrick CLI',
  tagline: 'One file. One command. A working dev environment — every time.',
  favicon: 'img/favicon.ico',

  url: 'https://salv4d.github.io',
  baseUrl: '/derrick/',

  organizationName: 'Salv4d',
  projectName: 'derrick',
  trailingSlash: false,

  markdown: {
    mermaid: true,
  },
  themes: [
    '@docusaurus/theme-mermaid',
    // Local, zero-config search. Free, no Algolia application required.
    // When/if we get DocSearch approval, swap this for @docusaurus/theme-search-algolia.
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      /** @type {import("@easyops-cn/docusaurus-search-local").PluginOptions} */
      ({
        hashed: true,
        indexBlog: false,
        docsRouteBasePath: '/',
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      }),
    ],
  ],

  onBrokenLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/',
          sidebarPath: './sidebars.js',
          editUrl: 'https://github.com/Salv4d/derrick/tree/main/',
          // Versioning: current docs/ is "next" (main); each release is
          // snapshotted into versioned_docs/version-X.Y.Z/. On a release:
          //   npm run docusaurus docs:version X.Y.Z
          // then update lastVersion below to make it the default.
          lastVersion: '0.3.0',
          includeCurrentVersion: true,
          versions: {
            current: {
              label: 'Next',
              path: 'next',
              banner: 'unreleased',
            },
            '0.3.0': {
              label: '0.3.0 (stable)',
              path: '',
            },
          },
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: {
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'Derrick',
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docs',
            position: 'left',
            label: 'Docs',
          },
          {
            to: '/why_derrick',
            label: 'Why Derrick?',
            position: 'left',
          },
          {
            to: '/use_cases/',
            label: 'Recipes',
            position: 'left',
          },
          {
            type: 'docsVersionDropdown',
            position: 'right',
          },
          {
            href: 'https://github.com/Salv4d/derrick',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Learn',
            items: [
              {label: 'Introduction', to: '/'},
              {label: 'Getting Started', to: '/getting_started'},
              {label: 'Why Derrick?', to: '/why_derrick'},
            ],
          },
          {
            title: 'Reference',
            items: [
              {label: 'CLI & Config', to: '/api_reference'},
              {label: 'Architecture', to: '/architecture'},
              {label: 'Glossary', to: '/glossary'},
            ],
          },
          {
            title: 'More',
            items: [
              {label: 'Troubleshooting', to: '/troubleshooting'},
              {label: 'Contributing', to: '/contributing'},
              {label: 'GitHub', href: 'https://github.com/Salv4d/derrick'},
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Derrick CLI. Built with Docusaurus.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
      },
    }),
};

export default config;
