const fs = require('fs');
const path = require('path');

const ROOT = path.join(__dirname, '..', '..', '..');
const REGISTRY_DIR = path.join(ROOT, 'registry');
const BACKEND_MODULES_DIR = path.join(ROOT, 'backend', 'internal', 'adapters', 'http', 'modules');
const CLI_TEMPLATES_DIR = path.join(ROOT, 'backend', 'cmd', 'kthulu-cli', 'templates');

function syncModules() {
  console.log('🔄 Syncing Modules from Backend...');
  if (!fs.existsSync(BACKEND_MODULES_DIR)) {
      console.warn('⚠️ Backend modules dir not found at:', BACKEND_MODULES_DIR);
      return;
  }

  const items = fs.readdirSync(BACKEND_MODULES_DIR);
  items.forEach(file => {
    if (file.endsWith('.go') && !file.endsWith('_test.go')) {
      const id = file.replace('.go', '');
      const reserved = ['module', 'routes', 'registry', 'static', 'builtin', 'module_set', 'shared_providers', 'access'];
      if (reserved.includes(id)) return;

      const itemDir = path.join(REGISTRY_DIR, 'modules', id);
      if (!fs.existsSync(itemDir)) fs.mkdirSync(itemDir, { recursive: true });

      const metadata = {
        id,
        name: id.charAt(0).toUpperCase() + id.slice(1) + ' Module',
        type: 'module',
        description: `Automatically discovered Kthulu module: ${id}`,
        author: 'Kthulu Core',
        stars: (id === 'auth' || id === 'ai') ? 100 : 0,
        icon: id === 'auth' ? 'Shield' : id === 'ai' ? 'Sparkles' : 'Box'
      };

      fs.writeFileSync(path.join(itemDir, 'metadata.json'), JSON.stringify(metadata, null, 2));
      
      const indexPath = path.join(itemDir, 'index.md');
      fs.writeFileSync(indexPath, `---
title: "${metadata.name}"
description: "${metadata.description}"
type: "module"
author: "${metadata.author}"
stars: ${metadata.stars}
icon: "${metadata.icon}"
---

# ${metadata.name}
`);
    }
  });
}

function syncStarters() {
  console.log('🔄 Syncing Starters from CLI Templates...');
  const starters = [
      { id: 'base-api', name: 'Base API', desc: 'The standard Kthulu API structure.', icon: 'Zap' },
      { id: 'saas-starter', name: 'SaaS Starter', desc: 'Full-stack SaaS with Auth and Billing.', icon: 'ShoppingBag' }
  ];

  starters.forEach(s => {
      const itemDir = path.join(REGISTRY_DIR, 'starters', s.id);
      if (!fs.existsSync(itemDir)) fs.mkdirSync(itemDir, { recursive: true });

      const metadata = {
        id: s.id,
        name: s.name,
        type: 'starter',
        description: s.desc,
        author: 'Kthulu Team',
        stars: 42,
        icon: s.icon
      };

      fs.writeFileSync(path.join(itemDir, 'metadata.json'), JSON.stringify(metadata, null, 2));
      
      // Starter Blueprint
      fs.writeFileSync(path.join(itemDir, 'blueprint.yaml'), `name: "${s.name}"\ntemplate: "${s.id}"\nmodules: ["auth", "ai"]\n`);

      const indexPath = path.join(itemDir, 'index.md');
      fs.writeFileSync(indexPath, `---
title: "${s.name}"
description: "${s.desc}"
type: "starter"
author: "Kthulu Team"
stars: 42
icon: "${s.icon}"
---

# ${s.name}
`);
  });
}

function syncPlugins() {
    console.log('🔄 Syncing Plugins...');
    const plugins = [
        { id: 'aws-deploy', name: 'AWS Deploy', desc: 'One-click deployment to AWS.', icon: 'Cloud' }
    ];

    plugins.forEach(p => {
        const itemDir = path.join(REGISTRY_DIR, 'plugins', p.id);
        if (!fs.existsSync(itemDir)) fs.mkdirSync(itemDir, { recursive: true });

        const metadata = {
          id: p.id,
          name: p.name,
          type: 'plugin',
          description: p.desc,
          author: 'CloudOps',
          stars: 88,
          icon: p.icon
        };

        fs.writeFileSync(path.join(itemDir, 'metadata.json'), JSON.stringify(metadata, null, 2));
        
        // Plugin Binaries Directory
        const binDir = path.join(itemDir, 'bin');
        if (!fs.existsSync(binDir)) fs.mkdirSync(binDir);
        fs.writeFileSync(path.join(binDir, '.keep'), '');

        const indexPath = path.join(itemDir, 'index.md');
        fs.writeFileSync(indexPath, `---
title: "${p.name}"
description: "${p.desc}"
type: "plugin"
author: "${metadata.author}"
stars: ${metadata.stars}
icon: "${metadata.icon}"
---

# ${p.name}
`);
    });
}

function main() {
  syncModules();
  syncStarters();
  syncPlugins();
  console.log('✅ Registry Sync Complete!');
}

main();
