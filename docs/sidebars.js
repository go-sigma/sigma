// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    'sigma',
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'quickstart',
      ],
    },
    'configuration',
    {
      type: 'category',
      label: 'Push to sigma',
      items: [
        'push/docker',
        'push/helm',
      ],
    },
  ],
};

module.exports = sidebars;
