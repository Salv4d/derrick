// @ts-check

/**
 * Explicit sidebar — ordered by reader journey, not filesystem.
 *
 * Order:
 *   1. Introduction        — what is Derrick, how to think about it
 *   2. Installation        — one-line install + all alternatives
 *   3. Getting Started     — first-project tutorial
 *   4. Why Derrick?        — honest comparison with alternatives
 *   5. Recipes             — copy-paste derrick.yaml for real projects
 *   6. CLI & Config        — reference for every command and yaml field
 *   7. Architecture        — Provider / State / Hooks mental model
 *   8. Troubleshooting     — common errors and fixes
 *   9. Contributing        — dev setup, tests, adding a provider
 *  10. Glossary            — term definitions
 *
 * Keep this ordering intentional. Each reader persona lands in a
 * different slot: evaluators hit (4), builders hit (3), operators
 * hit (6+8), contributors hit (9).
 */

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docs: [
    'intro',
    'installation',
    'getting_started',
    'why_derrick',
    {
      type: 'category',
      label: 'Recipes',
      link: {type: 'doc', id: 'use_cases/index'},
      items: [
        'use_cases/supabase',
        'use_cases/plausible',
        'use_cases/grafana',
        'use_cases/ghost',
        'use_cases/appwrite',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      collapsed: false,
      items: [
        'api_reference',
        'architecture',
        'glossary',
      ],
    },
    'troubleshooting',
    'contributing',
  ],
};

export default sidebars;
